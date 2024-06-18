package main

import (
	"database/sql"
	"log"
)

func initDb(DB *sql.DB) {

	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS casbin_rule (
		id VARCHAR(100) PRIMARY KEY,
		ptype VARCHAR(100),
		v0 VARCHAR(100),
		v1 VARCHAR(100),
		v2 VARCHAR(100),
		v3 VARCHAR(100)
	);`)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS domains (
		domain_id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL
	);`)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS users (
        user_id SERIAL PRIMARY KEY,
		domain_id INT NOT NULL,
        name VARCHAR(50) UNIQUE NOT NULL,
		FOREIGN KEY (domain_id) REFERENCES domains(domain_id)
    );`)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS roles (
		role_id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL
	);`)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS user_roles (
		id SERIAL PRIMARY KEY,
		domain_id INT NOT NULL,
		user_id INT NOT NULL,
		role_id INT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(user_id),
		FOREIGN KEY (role_id) REFERENCES roles(role_id),
		FOREIGN KEY (domain_id) REFERENCES domains(domain_id)
		);`)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

}
