package psql

import (
	"context"
	"time"

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
		Select("id", "number", "text", "thread_id", "sage", "op_marker", "browser_fingerprint", "ip", "created_at", "has_attachment", "deleted_at", "purged_at").
		From("post").
		Where(squirrel.Eq{"id": id}),
	)
	var post database.Post
	if err := row.Scan(&post.ID, &post.Number, &post.Text, &post.ThreadID, &post.Sage, &post.OpMarker, &post.BrowserFingerprint, &post.IP, &post.CreatedAt, &post.HasAttachment, &post.DeletedAt, &post.PurgedAt); err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repo) GetPosts(ctx context.Context, threadID uuid.UUID, includeDeleted bool) ([]database.Post, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "number", "text", "sage", "op_marker", "ip", "thread_id", "created_at", "browser_fingerprint", "has_attachment", "deleted_at", "purged_at").
		From("post").
		Where(squirrel.Eq{"thread_id": threadID}).
		Where(visibilityFilter(includeDeleted)).
		OrderBy("number ASC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Post, error) {
		var post database.Post
		err := row.Scan(&post.ID, &post.Number, &post.Text, &post.Sage, &post.OpMarker, &post.IP, &post.ThreadID, &post.CreatedAt, &post.BrowserFingerprint, &post.HasAttachment, &post.DeletedAt, &post.PurgedAt)
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

func (r *repo) GetBoardMetrics(ctx context.Context, from, to time.Time) ([]database.BoardMetrics, error) {
	rows, err := r.query(ctx, r.psqb.
		Select(
			"b.id as board_id",
			"b.alias as board_alias",
			"COUNT(p.id) as post_count",
			"COUNT(p.id) FILTER (WHERE p.deleted_at IS NOT NULL) as deleted_count",
			"COUNT(p.id) FILTER (WHERE p.sage = true) as sage_count",
			"COUNT(DISTINCT p.thread_id) as thread_count",
		).
		From("post p").
		Join("thread t ON p.thread_id = t.id").
		Join("board b ON t.board_id = b.id").
		Where(squirrel.GtOrEq{"p.created_at": from}).
		Where(squirrel.LtOrEq{"p.created_at": to}).
		GroupBy("b.id", "b.alias").
		OrderBy("b.alias ASC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.BoardMetrics, error) {
		var bm database.BoardMetrics
		err := row.Scan(&bm.BoardID, &bm.BoardAlias, &bm.PostCount, &bm.DeletedCount, &bm.SageCount, &bm.ThreadCount)
		return bm, err
	})
}

func (r *repo) GetRepliesForThread(ctx context.Context, threadID uuid.UUID, includeDeleted bool) (uint64, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("COUNT(id)").
		From("post").
		Where(squirrel.Eq{"thread_id": threadID}).
		Where(visibilityFilter(includeDeleted)),
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

// SoftDeletePost marks the post soft-deleted. Posts have no children so
// there's no cascade.
func (r *repo) SoftDeletePost(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.
		Update("post").
		Set("deleted_at", time.Now()).
		Where(squirrel.Eq{"id": id}),
	)
}

// RestorePost undoes the soft-delete on the post while still within the 24h
// grace window and it hasn't been purged.
func (r *repo) RestorePost(ctx context.Context, id uuid.UUID) error {
	const stmt = `
UPDATE post SET deleted_at = NULL
WHERE id = $1 AND deleted_at IS NOT NULL AND purged_at IS NULL AND deleted_at > now() - interval '24 hours'`
	_, err := r.q.Exec(ctx, stmt, id)
	return err
}

func (r *repo) ListPostsDueForPurge(ctx context.Context, before time.Time) ([]database.Post, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "number", "text", "sage", "op_marker", "ip", "thread_id", "created_at", "browser_fingerprint", "has_attachment", "deleted_at", "purged_at").
		From("post").
		Where(squirrel.Lt{"deleted_at": before}).
		Where(squirrel.Eq{"purged_at": nil}),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Post, error) {
		var post database.Post
		err := row.Scan(&post.ID, &post.Number, &post.Text, &post.Sage, &post.OpMarker, &post.IP, &post.ThreadID, &post.CreatedAt, &post.BrowserFingerprint, &post.HasAttachment, &post.DeletedAt, &post.PurgedAt)
		return post, err
	})
}

func (r *repo) MarkPostPurged(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.
		Update("post").
		Set("purged_at", time.Now()).
		Where(squirrel.Eq{"id": id}),
	)
}
