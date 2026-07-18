package ban

import (
	"context"
	"encoding/json"
	"errors"
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

func (s *redisStore) SetBan(ctx context.Context, key string, record Record, ttl time.Duration) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, data, ttl).Err()
}

func (s *redisStore) GetBan(ctx context.Context, key string) (*Record, error) {
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return &record, nil
}
