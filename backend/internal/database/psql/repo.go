package psql

import (
	"database/sql"
	"log"

	"github.com/Fodro/saberchan/internal/database"

	_ "github.com/lib/pq"
)

type repo struct {
	db *sql.DB
}

func NewRepo(connStr string) database.Repository {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return &repo{db: db}
}

func (r *repo) Ping() error {
	return r.db.Ping()
}