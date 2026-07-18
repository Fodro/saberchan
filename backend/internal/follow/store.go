package follow

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStore struct {
	rdb *redis.Client
}

// NewRedisStore builds a Store backed by rdb.
func NewRedisStore(rdb *redis.Client) Store {
	return &redisStore{rdb: rdb}
}

func (s *redisStore) Touch(ctx context.Context, key string, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, "1", ttl).Err()
}

func (s *redisStore) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *redisStore) ExistsMany(ctx context.Context, keys []string) ([]bool, error) {
	out := make([]bool, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Exists(ctx, key)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	for i, cmd := range cmds {
		n, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		out[i] = n > 0
	}
	return out, nil
}
