package psql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

func (r *repo) AddBoard(board *database.Board) error {
	stmt := `INSERT INTO board (id, alias, name, description, author, locked) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(stmt, board.ID, board.Alias, board.Name, board.Description, board.Author, board.Locked)
	return err
}

func (r *repo) DeleteBoard(id uuid.UUID) error {
	stmt := `DELETE FROM board WHERE id = $1 CASCADE`
	_, err := r.db.Exec(stmt, id)
	return err
}

func (r *repo) GetBoardByAlias(alias string) (*database.Board, error) {
	stmt := `SELECT id, alias, name, description, locked FROM board WHERE alias = $1`
	row := r.db.QueryRow(stmt, alias)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked); err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoardById(id uuid.UUID) (*database.Board, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("board id is required: %w", sql.ErrNoRows)
	}

	stmt := `SELECT id, alias, name, description, locked FROM board WHERE id = $1`
	row := r.db.QueryRow(stmt, id)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("board %s not found: %w", id, err)
		}
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoards() ([]database.Board, error) {
	stmt := `SELECT id, alias, name, description, locked FROM board ORDER BY created_at ASC`
	rows, err := r.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	return collectRows(rows, func(board *database.Board) error {
		return rows.Scan(&board.ID, &board.Alias, &board.Name, &board.Description, &board.Locked)
	})
}

func (r *repo) UpdateBoard(board *database.Board) error {
	stmt := `UPDATE board SET alias = $1, name = $2, description = $3 WHERE id = $4`
	_, err := r.db.Exec(stmt, board.Alias, board.Name, board.Description, board.ID)
	return err
}
