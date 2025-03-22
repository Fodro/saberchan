package psql

import (
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

func (r *repo) AddPost(post *database.Post) error {
	stmt := `INSERT INTO posts (id, text, thread_id, sage) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(stmt, post.ID, post.Text, post.ThreadID, post.Sage)
	return err
}

func (r *repo) DeletePost(id uuid.UUID) error {
	stmt := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.Exec(stmt, id)
	return err
}

func (r *repo) GetPost(id uuid.UUID) (*database.Post, error) {
	stmt := `SELECT id, number, text, thread_id, sage, op_marker, ip, created_at FROM posts WHERE id = $1`
	row := r.db.QueryRow(stmt, id)
	var post database.Post
	if err := row.Scan(&post.ID, &post.Number, &post.Text, &post.ThreadID, &post.Sage, &post.OpMarker, &post.IP, &post.CreatedAt); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repo) GetPosts(threadID uuid.UUID) ([]database.Post, error) {
	stmt := `SELECT id, number, text, sage, op_marker, ip, created_at FROM posts WHERE thread_id = $1 ORDER BY number ASC`
	rows, err := r.db.Query(stmt, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]database.Post, 0)
	for rows.Next() {
		var post database.Post
		if err := rows.Scan(&post.ID, &post.Number, &post.Text, &post.Sage, &post.OpMarker, &post.IP, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *repo) GetOPPost(threadID uuid.UUID) (*database.Post, error) {
	stmt := `SELECT id, number, text, thread_id, sage, op_marker, ip, created_at FROM posts WHERE thread_id = $1 ORDER_BY number ASC LIMIT 1`
	row := r.db.QueryRow(stmt, threadID)
	var post database.Post
	if err := row.Scan(&post.ID, &post.Number, &post.Text, &post.ThreadID, &post.Sage, &post.OpMarker, &post.IP, &post.CreatedAt); err != nil {
		return nil, err
	}
	return &post, nil
}