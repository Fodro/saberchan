package psql

import (
	"time"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

func (r *repo) AddThread(thread *database.Thread) error {
	stmt := `INSERT INTO thread (id, board_id, title) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(stmt, thread.ID, thread.BoardID, thread.Title)
	return err
}

func (r *repo) DeleteThread(id uuid.UUID) error {
	stmt := `DELETE FROM thread WHERE id = $1 CASCADE`
	_, err := r.db.Exec(stmt, id)
	return err
}

func (r *repo) GetThread(id uuid.UUID) (*database.Thread, error) {
	stmt := `SELECT id, board_id, title, locked, updated_at FROM thread WHERE id = $1`
	row := r.db.QueryRow(stmt, id)
	var thread database.Thread
	if err := row.Scan(&thread.ID, &thread.BoardID, &thread.Title, &thread.Locked, &thread.UpdatedAt); err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *repo) GetThreads(boardID uuid.UUID) ([]database.Thread, error) {
	stmt := `SELECT id, board_id, title, locked, updated_at FROM thread WHERE board_id = $1 ORDER BY updated_at DESC`
	rows, err := r.db.Query(stmt, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threads := make([]database.Thread, 0)
	for rows.Next() {
		var thread database.Thread
		if err := rows.Scan(&thread.ID, &thread.BoardID, &thread.Title, &thread.Locked, &thread.UpdatedAt); err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

func (r *repo) BumpThread(id uuid.UUID) error {
	stmt := `UPDATE thread SET updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(stmt, time.Now(), id)
	return err
}

func (r *repo) CheckIfThreadBelowBumpLimit(id uuid.UUID) (bool, error) {
	stmt := `SELECT t.id FROM thread t 
			JOIN post p on p.thread_id = t.id
			WHERE t.id = $1
			GROUP BY t.id
			HAVING count(p.id) < 500`
	if err := r.db.QueryRow(stmt, id).Scan(&id); err != nil {
		return false, nil
	}

	return true, nil
}
