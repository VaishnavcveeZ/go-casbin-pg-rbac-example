package main

type User struct {
	UserId int    `json:"user_id"`
	Name   string `json:"name"`
	Roles  []Role `json:"roles"`
}

type Role struct {
	RoleId int    `json:"role_id"`
	Role   string `json:"role"`
}

type Policy struct {
	RoleId   int    `json:"role_id"`
	Resource string `json:"resource"`
	Scope    string `json:"scope"`
}

type EnforcePermission struct {
	UserId   string `json:"user_id"`
	Resource string `json:"resource"`
	Scope    string `json:"scope"`
}

// type UserWithRole struct {
// 	UserId int    `json:"user_id"`
// 	Name   string `json:"name"`
// 	RoleId string `json:"role_id"`
// 	Role   string `json:"role"`
// }
