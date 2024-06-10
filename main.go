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
	// a, _ := pgadapter.NewAdapter("postgresql://username:password@postgres:5432/database?sslmode=disable") // Your driver and data source.
	// Alternatively, you can construct an adapter instance with *pg.Options:
	// a, _ := pgadapter.NewAdapter(&pg.Options{
	// 	Database: "casbin_example",
	// 	User:     "postgres",
	// 	Password: "mysecretpassword",
	// })

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

	// // Check the permission.
	// ok, err := e.Enforce("domain1", "alice", "data1", "read")
	// if err != nil {
	// 	log.Fatalf("err: %v", err)
	// }

	// if ok {
	// 	log.Println("hi")
	// } else {
	// 	log.Println("no hi")
	// }

}
