package psql

import (
	"context"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) GetFollowThreadInfos(ctx context.Context, ids []uuid.UUID) ([]database.FollowThreadInfo, error) {
	if len(ids) == 0 {
		return []database.FollowThreadInfo{}, nil
	}

	const stmt = `
SELECT t.id, t.title, b.alias,
  GREATEST(COUNT(p.id)::bigint - 1, 0) AS replies,
  COUNT(p.id) < 500 AS below_bump
FROM thread t
JOIN board b ON b.id = t.board_id
LEFT JOIN post p ON p.thread_id = t.id AND p.deleted_at IS NULL
WHERE t.id = ANY($1) AND t.deleted_at IS NULL AND t.purged_at IS NULL
GROUP BY t.id, t.title, b.alias`

	rows, err := r.q.Query(ctx, stmt, ids)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (database.FollowThreadInfo, error) {
		var info database.FollowThreadInfo
		err := row.Scan(&info.ID, &info.Title, &info.BoardAlias, &info.RepliesCount, &info.BelowBumpLimit)
		return info, err
	})
}
