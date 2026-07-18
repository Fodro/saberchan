package psql

import "database/sql"

// collectRows drains rows, always closes them, and surfaces rows.Err().
func collectRows[T any](rows *sql.Rows, scan func(*T) error) ([]T, error) {
	defer rows.Close()

	out := make([]T, 0)
	for rows.Next() {
		var item T
		if err := scan(&item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
