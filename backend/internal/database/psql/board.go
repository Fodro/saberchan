package psql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AddBoard(ctx context.Context, board *database.Board) error {
	return r.exec(ctx, r.psqb.
		Insert("board").
		Columns("id", "alias", "name", "description", "author", "locked").
		Values(board.ID, board.Alias, board.Name, board.Description, board.Author, board.Locked),
	)
}

func (r *repo) DeleteBoard(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.Delete("board").Where(squirrel.Eq{"id": id}))
}

// visibilityFilter returns the WHERE clause for the deleted/purged visibility
// rule: public callers only see live rows (deleted_at IS NULL); admin callers
// (includeDeleted) additionally see soft-deleted rows that haven't been purged.
func visibilityFilter(includeDeleted bool) squirrel.Sqlizer {
	if includeDeleted {
		return squirrel.Eq{"purged_at": nil}
	}
	return squirrel.Eq{"deleted_at": nil}
}

func (r *repo) GetBoardByAlias(ctx context.Context, alias string, includeDeleted bool) (*database.Board, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked", "deleted_at", "purged_at").
		From("board").
		Where(squirrel.Eq{"alias": alias}).
		Where(visibilityFilter(includeDeleted)),
	)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked, &board.DeletedAt, &board.PurgedAt); err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoardById(ctx context.Context, id uuid.UUID, includeDeleted bool) (*database.Board, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("board id is required: %w", pgx.ErrNoRows)
	}

	row := r.queryRow(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked", "deleted_at", "purged_at").
		From("board").
		Where(squirrel.Eq{"id": id}).
		Where(visibilityFilter(includeDeleted)),
	)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked, &board.DeletedAt, &board.PurgedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("board %s not found: %w", id, err)
		}
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoards(ctx context.Context, includeDeleted bool) ([]database.Board, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked", "deleted_at", "purged_at").
		From("board").
		Where(visibilityFilter(includeDeleted)).
		OrderBy("created_at ASC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Board, error) {
		var board database.Board
		err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked, &board.DeletedAt, &board.PurgedAt)
		return board, err
	})
}

func (r *repo) UpdateBoard(ctx context.Context, board *database.Board) error {
	return r.exec(ctx, r.psqb.
		Update("board").
		Set("alias", board.Alias).
		Set("name", board.Name).
		Set("description", board.Description).
		Where(squirrel.Eq{"id": board.ID}),
	)
}

// SoftDeleteBoard marks the board and cascades the soft-delete to its threads
// and posts in a single atomic statement. Cascaded children share the board's
// deleted_at timestamp so restore can match them exactly.
func (r *repo) SoftDeleteBoard(ctx context.Context, id uuid.UUID) error {
	const stmt = `
WITH b AS (
	UPDATE board SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL
	RETURNING id, deleted_at
),
t AS (
	UPDATE thread SET deleted_at = (SELECT deleted_at FROM b)
	WHERE board_id = (SELECT id FROM b) AND deleted_at IS NULL
	RETURNING id
)
UPDATE post SET deleted_at = (SELECT deleted_at FROM b)
WHERE thread_id IN (SELECT id FROM thread WHERE board_id = (SELECT id FROM b))
	AND deleted_at IS NULL`
	_, err := r.q.Exec(ctx, stmt, id)
	return err
}

// RestoreBoard undoes the soft-delete on the board (within the 24h grace
// window) and cascades restore only to threads/posts whose deleted_at is
// identical to the board's — previously independently deleted children are left alone.
func (r *repo) RestoreBoard(ctx context.Context, id uuid.UUID) error {
	const stmt = `
WITH old AS (
	SELECT id, deleted_at
	FROM board
	WHERE id = $1
		AND deleted_at IS NOT NULL
		AND purged_at IS NULL
		AND deleted_at > now() - interval '24 hours'
),
b AS (
	UPDATE board SET deleted_at = NULL
	WHERE id IN (SELECT id FROM old)
	RETURNING id
),
t AS (
	UPDATE thread SET deleted_at = NULL
	WHERE board_id IN (SELECT id FROM old)
		AND purged_at IS NULL
		AND deleted_at IS NOT NULL
		AND deleted_at = (SELECT deleted_at FROM old)
	RETURNING id
)
UPDATE post SET deleted_at = NULL
WHERE thread_id IN (SELECT id FROM thread WHERE board_id IN (SELECT id FROM old))
	AND purged_at IS NULL
	AND deleted_at IS NOT NULL
	AND deleted_at = (SELECT deleted_at FROM old)`
	_, err := r.q.Exec(ctx, stmt, id)
	return err
}

func (r *repo) ListBoardsDueForPurge(ctx context.Context, before time.Time) ([]database.Board, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked", "deleted_at", "purged_at").
		From("board").
		Where(squirrel.Lt{"deleted_at": before}).
		Where(squirrel.Eq{"purged_at": nil}),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Board, error) {
		var board database.Board
		err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked, &board.DeletedAt, &board.PurgedAt)
		return board, err
	})
}

func (r *repo) MarkBoardPurged(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, r.psqb.
		Update("board").
		Set("purged_at", time.Now()).
		Where(squirrel.Eq{"id": id}),
	)
}
