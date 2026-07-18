package psql

import (
	"context"
	"errors"
	"fmt"

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

func (r *repo) GetBoardByAlias(ctx context.Context, alias string) (*database.Board, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked").
		From("board").
		Where(squirrel.Eq{"alias": alias}),
	)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked); err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoardById(ctx context.Context, id uuid.UUID) (*database.Board, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("board id is required: %w", pgx.ErrNoRows)
	}

	row := r.queryRow(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked").
		From("board").
		Where(squirrel.Eq{"id": id}),
	)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("board %s not found: %w", id, err)
		}
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoards(ctx context.Context) ([]database.Board, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("id", "alias", "name", "description", "locked").
		From("board").
		OrderBy("created_at ASC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Board, error) {
		var board database.Board
		err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked)
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
