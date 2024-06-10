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

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Policy struct {
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Scope    string `json:"scope"`
}

type EnforcePermission struct {
	UserId   string `json:"user_id"`
	Resource string `json:"resource"`
	Scope    string `json:"scope"`
}

func AddUserHandler(c *gin.Context) {

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert user into database
	_, err := DB.Exec("INSERT INTO users (name, role) VALUES ($1, $2)", user.Name, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
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

		ok, err := E.AddPolicy(p.Role, c.Request.Host, p.Resource, p.Scope)
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

func GetUsersHandler(c *gin.Context) {

	rows, err := DB.Query("SELECT id, name, role from users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Iterate over the rows and build the user list with permissions
	var userListWithPermissions []map[string]interface{}
	for rows.Next() {
		var id int
		var name, role string
		if err := rows.Scan(&id, &name, &role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user := map[string]interface{}{
			"id":   id,
			"name": name,
			"role": role,
		}
		userListWithPermissions = append(userListWithPermissions, user)
	}

	c.JSON(http.StatusOK, userListWithPermissions)
}

func EnforceUserPermissionHandler(c *gin.Context) {

	var enforcePermission EnforcePermission
	if err := c.ShouldBindJSON(&enforcePermission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User

	query := "SELECT id, name, role FROM users WHERE id = $1"
	row := DB.QueryRow(query, enforcePermission.UserId)
	err := row.Scan(&user.Id, &user.Name, &user.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check the permission.
	ok, err := E.Enforce(user.Role, c.Request.Host, enforcePermission.Resource, enforcePermission.Scope)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ok {
		c.JSON(http.StatusOK, user)

	} else {
		c.JSON(http.StatusUnauthorized, user)
	}
}
