package main

import (
	"github.com/casbin/casbin/v2"
)

var E *casbin.Enforcer

func InitCasbin() error {

	var err error
	E, err = casbin.NewEnforcer("rbac_model.conf", A)
	if err != nil {
		return err
	}

	// _, err = E.AddGroupingPolicy("teacher", "student") // teacher inherits all permissions of student
	// if err != nil {
	// 	return err
	// }
	// _, err = E.AddGroupingPolicy("admin", "teacher") // admin inherits all permissions of teacher
	// if err != nil {
	// 	return err
	// }

	// Load the policy from DB.
	err = E.LoadPolicy()
	if err != nil {
		return err
	}

	return nil
}
