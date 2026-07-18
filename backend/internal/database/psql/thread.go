package psql

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AddThread(ctx context.Context, thread *database.Thread) error {
	return r.exec(ctx, r.psqb.
		Insert("thread").
		Columns("id", "board_id", "title").
		Values(thread.ID, thread.BoardID, thread.Title),
	)
}

func (r *repo) DeleteThread(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.Delete("thread").Where(squirrel.Eq{"id": id}))
}

func (r *repo) GetThread(ctx context.Context, id uuid.UUID) (*database.Thread, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("id", "board_id", "title", "locked", "updated_at").
		From("thread").
		Where(squirrel.Eq{"id": id}),
	)
	var thread database.Thread
	if err := row.Scan(&thread.ID, &thread.BoardID, &thread.Title, &thread.Locked, &thread.UpdatedAt); err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *repo) GetThreads(ctx context.Context, boardID uuid.UUID) ([]database.Thread, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "board_id", "title", "locked", "updated_at").
		From("thread").
		Where(squirrel.Eq{"board_id": boardID}).
		OrderBy("updated_at DESC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Thread, error) {
		var thread database.Thread
		err := row.Scan(&thread.ID, &thread.BoardID, &thread.Title, &thread.Locked, &thread.UpdatedAt)
		return thread, err
	})
}

func (r *repo) CountThreads(ctx context.Context, boardID uuid.UUID) (uint64, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("COUNT(*)").
		From("thread").
		Where(squirrel.Eq{"board_id": boardID}),
	)
	var n uint64
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// GetBoardCatalog keeps the multi-CTE query as raw SQL — DISTINCT ON + CTEs
// are a poor fit for squirrel and easier to keep correct this way.
func (r *repo) GetBoardCatalog(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]database.CatalogThread, error) {
	const stmt = `
WITH page AS (
	SELECT id, board_id, title, locked, updated_at
	FROM thread
	WHERE board_id = $1
	ORDER BY updated_at DESC
	LIMIT $2 OFFSET $3
),
op AS (
	SELECT DISTINCT ON (p.thread_id)
		p.id, p.thread_id, p.number, p.text, p.sage, p.op_marker, p.ip,
		p.created_at, p.browser_fingerprint, p.has_attachment
	FROM post p
	INNER JOIN page t ON t.id = p.thread_id
	ORDER BY p.thread_id, p.number ASC
),
counts AS (
	SELECT p.thread_id, GREATEST(COUNT(*)::bigint - 1, 0) AS replies
	FROM post p
	INNER JOIN page t ON t.id = p.thread_id
	GROUP BY p.thread_id
)
SELECT
	t.id, t.board_id, t.title, t.locked, t.updated_at,
	o.id, o.number, o.text, o.sage, o.op_marker, o.ip, o.created_at, o.browser_fingerprint, o.has_attachment,
	COALESCE(c.replies, 0)
FROM page t
INNER JOIN op o ON o.thread_id = t.id
LEFT JOIN counts c ON c.thread_id = t.id
ORDER BY t.updated_at DESC`

	rows, err := r.q.Query(ctx, stmt, boardID, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.CatalogThread, error) {
		var ct database.CatalogThread
		err := row.Scan(
			&ct.ID, &ct.BoardID, &ct.Title, &ct.Locked, &ct.UpdatedAt,
			&ct.OP.ID, &ct.OP.Number, &ct.OP.Text, &ct.OP.Sage, &ct.OP.OpMarker, &ct.OP.IP,
			&ct.OP.CreatedAt, &ct.OP.BrowserFingerprint, &ct.OP.HasAttachment,
			&ct.RepliesCount,
		)
		if err != nil {
			return ct, err
		}
		ct.OP.ThreadID = ct.ID
		return ct, nil
	})
}

func (r *repo) BumpThread(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.
		Update("thread").
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id}),
	)
}

func (r *repo) CheckIfThreadBelowBumpLimit(ctx context.Context, id uuid.UUID) (bool, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("t.id").
		From("thread t").
		Join("post p ON p.thread_id = t.id").
		Where(squirrel.Eq{"t.id": id}).
		GroupBy("t.id").
		Having("count(p.id) < 500"),
	)
	var scanned uuid.UUID
	if err := row.Scan(&scanned); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, nil
	}
	return true, nil
}
