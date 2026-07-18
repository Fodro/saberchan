package captcha

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -destination=mocks/mock_token_store.go -package=mocks github.com/Fodro/saberchan/internal/captcha TokenStore

// TokenStore persists captcha answers keyed by token.
type TokenStore interface {
	Set(ctx context.Context, token, answer string, ttl time.Duration) error
	GetDel(ctx context.Context, token string) (string, error)
}

type redisStore struct {
	rdb *redis.Client
}

func (s *redisStore) Set(ctx context.Context, token, answer string, ttl time.Duration) error {
	return s.rdb.Set(ctx, token, answer, ttl).Err()
}

func (s *redisStore) GetDel(ctx context.Context, token string) (string, error) {
	return s.rdb.GetDel(ctx, token).Result()
}
