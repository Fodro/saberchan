package health

import (
	"context"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/redis/go-redis/v9"
)

type service struct {
	repo database.Repository
	rdb  *redis.Client // optional; when set, readiness also pings Redis
}

func (s *service) Readiness(ctx context.Context) error {
	if err := s.repo.Ping(ctx); err != nil {
		return err
	}
	if s.rdb != nil {
		return s.rdb.Ping(ctx).Err()
	}
	return nil
}

func NewService(repo database.Repository) Service {
	return &service{repo: repo}
}

// NewServiceWithRedis also requires Redis to be reachable for readiness.
func NewServiceWithRedis(repo database.Repository, rdb *redis.Client) Service {
	return &service{repo: repo, rdb: rdb}
}
