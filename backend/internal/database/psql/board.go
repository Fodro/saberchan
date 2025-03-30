package psql

import (
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

func (r *repo) AddBoard(board *database.Board) error {
	stmt := `INSERT INTO boards (id, alias, name, description, author) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(stmt, board.ID, board.Alias, board.Name, board.Description, board.Author)
	return err
}

func (r *repo) DeleteBoard(id uuid.UUID) error {
	stmt := `DELETE FROM board WHERE id = $1 CASCADE`
	_, err := r.db.Exec(stmt, id)
	return err
}

func (r *repo) GetBoardByAlias(alias string) (*database.Board, error) {
	stmt := `SELECT id, alias, name, description FROM board WHERE alias = $1`
	row := r.db.QueryRow(stmt, alias)
	var board database.Board
	if err := row.Scan(&board.ID, &board.Alias, &board.Name, &board.Description); err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *repo) GetBoards() ([]database.Board, error) {
	stmt := `SELECT id, alias, name, description FROM board ORDER BY created_at ASC`
	rows, err := r.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boards := make([]database.Board, 0)
	for rows.Next() {
		var board database.Board
		if err := rows.Scan(&board.ID, &board.Alias, &board.Name, &board.Description); err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, nil
}

func (r *repo) UpdateBoard(board *database.Board) error {
	stmt := `UPDATE board SET alias = $1, name = $2, description = $3 WHERE id = $4`
	_, err := r.db.Exec(stmt, board.Alias, board.Name, board.Description, board.ID)
	return err
}
