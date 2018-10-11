package queries

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // postgres driver
)

var db *sql.DB

func InitDB(address, database string) *sql.DB {
	var err error
	db, err = sql.Open("postgres",
		"postgres://"+address+"/"+database+"?sslmode=disable")
	if err != nil {
		log.Panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	log.Printf("Successfully connected to %v, database %v\n", address, database)

	return db
}
