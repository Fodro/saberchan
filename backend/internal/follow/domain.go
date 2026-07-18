package follow

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const TTL = 7 * 24 * time.Hour

var ErrDead = errors.New("follow: thread is dead")

//go:generate mockgen -destination=mocks/mock_store.go -package=mocks github.com/Fodro/saberchan/internal/follow Store

// Store persists follow markers keyed by follow:thread:{uuid}.
type Store interface {
	Touch(ctx context.Context, key string, ttl time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	ExistsMany(ctx context.Context, keys []string) ([]bool, error)
}

// Status is a follow listing row for a thread.
type Status struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	BoardAlias   string    `json:"board_alias"`
	RepliesCount uint64    `json:"replies_count"`
	Dead         bool      `json:"dead"`
}

// Service manages thread follow markers and status.
type Service interface {
	// Follow sets/refreshes the Redis key only if the thread is below the bump
	// limit. Returns ErrDead when the thread is missing or bump-limited.
	Follow(ctx context.Context, threadID uuid.UUID) error
	// RefreshOnBump refreshes the TTL when the key already exists and the
	// thread is still below the bump limit; otherwise it is a no-op.
	RefreshOnBump(ctx context.Context, threadID uuid.UUID) error
	Status(ctx context.Context, ids []uuid.UUID) ([]Status, error)
}

// ThreadKey returns the Redis key for a followed thread.
func ThreadKey(id uuid.UUID) string {
	return "follow:thread:" + id.String()
}
