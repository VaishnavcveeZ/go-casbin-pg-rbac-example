package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Start() {

	// Initialize Gin router
	router := gin.Default()

	/*
		curl --location 'http://localhost:8080/user' \
		--header 'Content-Type: application/json' \
		--data '{
			"name": "chris",
			"role": "admin"
		}'
	*/
	router.POST("/user", AddUserHandler)

	/*
		curl --location 'http://localhost:8080/user' \
		--data ''
	*/
	router.GET("/user", GetUsersHandler)

	/*
		curl --location 'http://localhost:8080/policy' \
		--header 'Content-Type: application/json' \
		--data '[
			{
				"role": "admin",
				"resource": "dashboard",
				"scope": "write"
			}
		]'
	*/
	router.POST("/policy", AddPolicyHandler)

	/*
		curl --location --request DELETE 'http://localhost:8080/policy' \
		--header 'Content-Type: application/json' \
		--data '[
			{
				"role": "admin",
				"resource": "dashboard",
				"scope": "write"
			}
		]'
	*/
	router.DELETE("/policy", RemovePolicyHandler)

	/*
		curl --location 'http://localhost:8080/policy' \
		--data ''
	*/
	router.GET("/policy", GetAllPolicyHandler)

	/*
		curl --location 'http://localhost:8080/enforce/user/policy' \
		--header 'Content-Type: application/json' \
		--data '{
			"user_id": "1",
			"resource":"dashboard",
			"scope":"write"
		}'
	*/
	router.POST("/enforce/user/policy", EnforceUserPermissionHandler)

	// Run server
	router.Run(":8080")
}

func AddPolicyHandler(c *gin.Context) {
	var policy []Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if policy == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy list"})
		return
	}

	for _, p := range policy {

		if !ValidateRole(p.RoleId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid role %v", p.RoleId)})
			return
		}

		ok, err := E.AddPolicy("1", c.Request.Host, p.Resource, p.Scope)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to add policy: %+v", p)})
			return
		}
	}

	// Save the policy back to DB.
	err := E.SavePolicy()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func RemovePolicyHandler(c *gin.Context) {

	var policy []Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if policy == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy list"})
		return
	}

	for _, p := range policy {

		if !ValidateRole(p.RoleId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid role %v", p.RoleId)})
			return
		}

		ok, err := E.RemovePolicy(p.RoleId, c.Request.Host, p.Resource, p.Scope)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to remove policy: %+v", p)})
			return
		}
	}

	// Save the policy back to DB.
	err := E.SavePolicy()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func GetAllPolicyHandler(c *gin.Context) {

	p, err := E.GetPolicy()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"policy": p,
	})
}

func AddUserHandler(c *gin.Context) {

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, r := range user.Roles {

		if !ValidateRole(r.RoleId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid role %v", r)})
			return
		}
	}

	// Insert user into database
	res, err := DB.Exec("INSERT INTO users (name, domain_id) VALUES ($1, $2)", user.Name, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userId, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, r := range user.Roles {

		_, err = E.AddGroupingPolicy(userId, r.RoleId, c.Request.Host)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.Status(http.StatusOK)
}

func GetUsersHandler(c *gin.Context) {

	rows, err := DB.Query(`SELECT u.user_id, u.name, r.role_id, r.name
	FROM users u
	JOIN user_roles ur ON u.id = ur.user_id
	JOIN roles r ON ur.role_id = r.id
	JOIN domains d ON ur.domain_id = d.id
	WHERE d.name = $2`, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Iterate over the rows and build the user list with permissions
	userHash := make(map[int]User)

	for rows.Next() {
		var userId, roleId int
		var name, roleName string
		if err := rows.Scan(&userId, &name, &roleId, &roleName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		u, ok := userHash[userId]
		if ok {
			u.Roles = append(u.Roles, Role{RoleId: roleId, Role: roleName})
		} else {

			userHash[userId] = User{
				UserId: userId,
				Name:   name,
				Roles: []Role{{
					RoleId: roleId,
					Role:   roleName,
				}},
			}
		}
	}

	var userListWithRoles []User
	for _, v := range userHash {
		userListWithRoles = append(userListWithRoles, v)
	}

	c.JSON(http.StatusOK, userListWithRoles)
}

func EnforceUserPermissionHandler(c *gin.Context) {

	var enforcePermission EnforcePermission
	if err := c.ShouldBindJSON(&enforcePermission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
	SELECT u.user_id, u.name, r.role_id, r.name
	FROM users u
	JOIN user_roles ur ON u.id = ur.user_id
	JOIN roles r ON ur.role_id = r.id
	JOIN domains d ON ur.domain_id = d.id
	WHERE u.user_id = $1
	  AND d.name = $2
`

	// query := "SELECT id, name, role FROM users WHERE id = $1"
	rows, err := DB.Query(query, enforcePermission.UserId, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var u User
	for rows.Next() {

		var roleId int
		var roleName string

		err := rows.Scan(&u.UserId, &u.Name, &roleId, &roleName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		u.Roles = append(u.Roles, Role{RoleId: roleId, Role: roleName})

	}

	for _, role := range u.Roles {

		ok, err := E.Enforce(role, c.Request.Host, enforcePermission.Resource, enforcePermission.Scope)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if ok {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, user)
}
