package psql

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is implemented by *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type repo struct {
	pool *pgxpool.Pool
	q    DBTX
	psqb squirrel.StatementBuilderType
}

func NewRepo(pool *pgxpool.Pool) database.Repository {
	return &repo{
		pool: pool,
		q:    pool,
		psqb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *repo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *repo) InTx(ctx context.Context, fn func(tx database.Repository) error) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&repo{pool: r.pool, q: tx, psqb: r.psqb})
	})
}

func (r *repo) exec(ctx context.Context, qb squirrel.Sqlizer) error {
	sql, args, err := qb.ToSql()
	if err != nil {
		return err
	}
	_, err = r.q.Exec(ctx, sql, args...)
	return err
}

func (r *repo) query(ctx context.Context, qb squirrel.Sqlizer) (pgx.Rows, error) {
	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, err
	}
	return r.q.Query(ctx, sql, args...)
}

func (r *repo) queryRow(ctx context.Context, qb squirrel.Sqlizer) pgx.Row {
	sql, args, err := qb.ToSql()
	if err != nil {
		return errRow{err: err}
	}
	return r.q.QueryRow(ctx, sql, args...)
}

type errRow struct{ err error }

func (e errRow) Scan(_ ...any) error { return e.err }
