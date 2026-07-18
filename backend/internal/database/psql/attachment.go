package psql

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AddAttachment(ctx context.Context, attachment *database.Attachment) error {
	return r.exec(ctx, r.psqb.
		Insert("attachment").
		Columns("id", "post_id", "link", "name", "type", "key").
		Values(attachment.ID, attachment.PostID, attachment.Link, attachment.Name, attachment.Type, attachment.Key),
	)
}

func (r *repo) GetAttachments(ctx context.Context, postID uuid.UUID) ([]database.Attachment, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "link", "name", "type", "key").
		From("attachment").
		Where(squirrel.Eq{"post_id": postID}),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Attachment, error) {
		var a database.Attachment
		err := row.Scan(&a.ID, &a.Link, &a.Name, &a.Type, &a.Key)
		return a, err
	})
}

func (r *repo) GetAttachmentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]database.Attachment, error) {
	if len(postIDs) == 0 {
		return []database.Attachment{}, nil
	}
	rows, err := r.query(ctx, r.psqb.
		Select("id", "post_id", "link", "name", "type", "key").
		From("attachment").
		Where(squirrel.Eq{"post_id": postIDs}),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Attachment, error) {
		var a database.Attachment
		err := row.Scan(&a.ID, &a.PostID, &a.Link, &a.Name, &a.Type, &a.Key)
		return a, err
	})
}

func (r *repo) DeleteAttachmentsByPostID(ctx context.Context, postID uuid.UUID) error {
	return r.exec(ctx, r.psqb.Delete("attachment").Where(squirrel.Eq{"post_id": postID}))
}
