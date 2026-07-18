package psql

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AddConfig(ctx context.Context, config *database.Config) error {
	return r.exec(ctx, r.psqb.
		Insert("config").
		Columns("nickname", "bump_limit", "current", "site_title").
		Values(config.Nickname, config.BumpLimit, false, config.SiteName),
	)
}

func (r *repo) ChangeCurrConfig(ctx context.Context, configId uuid.UUID) error {
	apply := func(tr *repo) error {
		if err := tr.exec(ctx, tr.psqb.
			Update("config").
			Set("current", false).
			Where(squirrel.Eq{"current": true}),
		); err != nil {
			return err
		}
		return tr.exec(ctx, tr.psqb.
			Update("config").
			Set("current", true).
			Where(squirrel.Eq{"id": configId}),
		)
	}
	if _, ok := r.q.(pgx.Tx); ok {
		return apply(r)
	}
	return r.InTx(ctx, func(tx database.Repository) error {
		return apply(tx.(*repo))
	})
}

func (r *repo) GetCurrentConfig(ctx context.Context) (*database.Config, error) {
	row := r.queryRow(ctx, r.psqb.
		Select("nickname", "bump_limit", "site_title").
		From("config").
		Where(squirrel.Eq{"current": true}).
		OrderBy("created_at DESC").
		Limit(1),
	)
	var config database.Config
	if err := row.Scan(&config.Nickname, &config.BumpLimit, &config.SiteName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNoCurConfig
		}
		return nil, err
	}
	return &config, nil
}

func (r *repo) GetConfigs(ctx context.Context) ([]database.Config, error) {
	rows, err := r.query(ctx, r.psqb.
		Select("nickname", "bump_limit", "site_title").
		From("config").
		OrderBy("created_at DESC"),
	)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.Config, error) {
		var config database.Config
		err := row.Scan(&config.Nickname, &config.BumpLimit, &config.SiteName)
		return config, err
	})
}
