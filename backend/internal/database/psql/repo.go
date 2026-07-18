package psql

import (
	"database/sql"
	"log"
	"time"

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

	// Tuned for external clustered Postgres (compose/local still fine with these caps).
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &repo{db: db}
}

func (r *repo) Ping() error {
	return r.db.Ping()
}
