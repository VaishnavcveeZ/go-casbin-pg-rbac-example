package main

import (
	"database/sql"
	"fmt"
	"log"

	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/go-pg/pg/v10"
	_ "github.com/lib/pq"
)

var A *pgadapter.Adapter

var DB *sql.DB

const (
	host     = "127.0.0.1"
	port     = 5432
	user     = "postgres"
	password = "mysecretpassword"
	dbname   = "casbin_two"
)

func main() {

	var err error

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	opts, err := pg.ParseURL(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	db := pg.Connect(opts)
	defer db.Close()

	A, err = pgadapter.NewAdapterByDB(db, pgadapter.WithTableName("casbin_rule"))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	initDb(DB)

	// Initialize Casbin
	err = InitCasbin()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	Start()

}
