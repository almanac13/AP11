package repository

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewDB(conn string) *sql.DB {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
