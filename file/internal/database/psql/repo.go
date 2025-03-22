package psql

import (
	"database/sql"
	"log"

	"github.com/Fodro/saberchan/file/internal/database"
	"github.com/google/uuid"
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

func (r *repo) AddFile(f *database.File) error {
	stmt := `INSERT INTO file (id, post_id, key) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(stmt, f.ID, f.PostID, f.Key)
	return err
}

func (r *repo) DeleteFilesForPost(postID uuid.UUID) error {
	stmt := "UPDATE file SET deleted_at = current_timestamp WHERE post_id = $1"
	_, err := r.db.Exec(stmt, postID)
	return err
}

func (r *repo) GetFilesForPost(postID uuid.UUID) ([]*database.File, error) {
	stmt := "SELECT * FROM file WHERE post_id = $1 AND deleted_at IS NULL"
	rows, err := r.db.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*database.File
	for rows.Next() {
		f := &database.File{}
		err := rows.Scan(&f.ID, &f.PostID, &f.Key, &f.CreatedAt, &f.DeletedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (r *repo) Ping() error {
	return r.db.Ping()
}
