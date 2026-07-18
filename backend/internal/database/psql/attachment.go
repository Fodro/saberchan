package psql

import (
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

func (r *repo) AddAttachment(attachment *database.Attachment) error {
	stmt := `INSERT INTO attachment (id, post_id, link, name, type) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(stmt, attachment.ID, attachment.PostID, attachment.Link, attachment.Name, attachment.Type)
	return err
}

func (r *repo) GetAttachments(postID uuid.UUID) ([]database.Attachment, error) {
	stmt := `SELECT id, link, name, type FROM attachment WHERE post_id = $1`
	rows, err := r.db.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	return collectRows(rows, func(attachment *database.Attachment) error {
		return rows.Scan(&attachment.ID, &attachment.Link, &attachment.Name, &attachment.Type)
	})
}
