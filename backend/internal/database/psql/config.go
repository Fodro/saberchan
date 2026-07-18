package psql

import (
	"database/sql"
	"errors"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

func (r *repo) AddConfig(config *database.Config) error {
	stmt := `INSERT INTO config (nickname, bump_limit, current, site_title) VALUES ($1, $2, false, $3)`
	_, err := r.db.Exec(stmt, config.Nickname, config.BumpLimit, config.SiteName)
	return err
}

func (r *repo) ChangeCurrConfig(configId uuid.UUID) error {
	stmt := `UPDATE config SET current = false WHERE current = true`
	_, err := r.db.Exec(stmt)
	if err != nil {
		return err
	}

	stmt = `UPDATE config SET current = true WHERE id = $1`
	_, err = r.db.Exec(stmt, configId)
	return err
}

func (r *repo) GetCurrentConfig() (*database.Config, error) {
	stmt := `SELECT nickname, bump_limit, site_title FROM config WHERE current = true ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRow(stmt)
	var config database.Config
	if err := row.Scan(&config.Nickname, &config.BumpLimit, &config.SiteName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNoCurConfig
		}
		return nil, err
	}

	return &config, nil
}

func (r *repo) GetConfigs() ([]database.Config, error) {
	stmt := `SELECT nickname, bump_limit, site_title FROM config ORDER BY created_at DESC`
	rows, err := r.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	return collectRows(rows, func(config *database.Config) error {
		return rows.Scan(&config.Nickname, &config.BumpLimit, &config.SiteName)
	})
}
