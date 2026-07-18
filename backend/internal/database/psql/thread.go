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
	return collectRows(rows, func(thread *database.Thread) error {
		return rows.Scan(&thread.ID, &thread.BoardID, &thread.Title, &thread.Locked, &thread.UpdatedAt)
	})
}

func (r *repo) CountThreads(boardID uuid.UUID) (uint64, error) {
	stmt := `SELECT COUNT(*) FROM thread WHERE board_id = $1`
	var n uint64
	if err := r.db.QueryRow(stmt, boardID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *repo) GetBoardCatalog(boardID uuid.UUID, limit, offset int) ([]database.CatalogThread, error) {
	stmt := `
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

	rows, err := r.db.Query(stmt, boardID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]database.CatalogThread, 0)
	for rows.Next() {
		var ct database.CatalogThread
		if err := rows.Scan(
			&ct.ID, &ct.BoardID, &ct.Title, &ct.Locked, &ct.UpdatedAt,
			&ct.OP.ID, &ct.OP.Number, &ct.OP.Text, &ct.OP.Sage, &ct.OP.OpMarker, &ct.OP.IP,
			&ct.OP.CreatedAt, &ct.OP.BrowserFingerprint, &ct.OP.HasAttachment,
			&ct.RepliesCount,
		); err != nil {
			return nil, err
		}
		ct.OP.ThreadID = ct.ID
		out = append(out, ct)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
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
