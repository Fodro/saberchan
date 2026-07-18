package psql

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AddPost(ctx context.Context, post *database.Post) error {
	return r.exec(ctx, r.psqb.
		Insert("post").
		Columns("id", "text", "thread_id", "sage", "browser_fingerprint", "ip", "op_marker", "has_attachment").
		Values(post.ID, post.Text, post.ThreadID, post.Sage, post.BrowserFingerprint, post.IP, post.OpMarker, post.HasAttachment),
	)
}

func (r *repo) DeletePost(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.Delete("post").Where(squirrel.Eq{"id": id}))
}

func (r *repo) GetPost(ctx context.Context, id uuid.UUID) (*database.Post, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("id", "number", "text", "thread_id", "sage", "op_marker", "ip", "created_at", "has_attachment").
		From("post").
		Where(squirrel.Eq{"id": id}),
	)
	var post database.Post
	if err := row.Scan(&post.ID, &post.Number, &post.Text, &post.ThreadID, &post.Sage, &post.OpMarker, &post.IP, &post.CreatedAt, &post.HasAttachment); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repo) GetPosts(ctx context.Context, threadID uuid.UUID) ([]database.Post, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "number", "text", "sage", "op_marker", "ip", "thread_id", "created_at", "browser_fingerprint", "has_attachment").
		From("post").
		Where(squirrel.Eq{"thread_id": threadID}).
		OrderBy("number ASC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Post, error) {
		var post database.Post
		err := row.Scan(&post.ID, &post.Number, &post.Text, &post.Sage, &post.OpMarker, &post.IP, &post.ThreadID, &post.CreatedAt, &post.BrowserFingerprint, &post.HasAttachment)
		return post, err
	})
}

func (r *repo) GetOPPost(ctx context.Context, threadID uuid.UUID) (*database.Post, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("id", "number", "text", "thread_id", "sage", "op_marker", "ip", "created_at", "browser_fingerprint", "has_attachment").
		From("post").
		Where(squirrel.Eq{"thread_id": threadID}).
		OrderBy("number ASC").
		Limit(1),
	)
	var post database.Post
	if err := row.Scan(&post.ID, &post.Number, &post.Text, &post.ThreadID, &post.Sage, &post.OpMarker, &post.IP, &post.CreatedAt, &post.BrowserFingerprint, &post.HasAttachment); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repo) GetRepliesForThread(ctx context.Context, threadID uuid.UUID) (uint64, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("COUNT(id)").
		From("post").
		Where(squirrel.Eq{"thread_id": threadID}),
	)
	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}
	return count - 1, nil
}
