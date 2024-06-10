package main

import (
	"database/sql"
	"fmt"
	"log"

	pgadapter "github.com/casbin/casbin-pg-adapter"
	_ "github.com/lib/pq"
)

var A *pgadapter.Adapter

var DB *sql.DB

const (
	host     = "127.0.0.1"
	port     = 5432
	user     = "postgres"
	password = "mysecretpassword"
	dbname   = "casbin"
)

func main() {

	var err error
	A, err = pgadapter.NewAdapter("postgresql://postgres:mysecretpassword@localhost:5432/casbin?sslmode=disable")
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50) UNIQUE NOT NULL,
        role varchar(50) NOT NULL
    );`)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS casbin_rule (
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

	// Initialize Casbin
	err = InitCasbin()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	Start()

}
