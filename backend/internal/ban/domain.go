// Package ban implements IP/fingerprint bans backed by Redis. A ban record
// is stored under ban:ip:{ip} and/or ban:fp:{fp} with a TTL equal to the
// remaining ban duration; permanent bans use a 100-year TTL since Redis has
// no "never expires" option that still carries a value.
package ban

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_store.go -package=mocks github.com/Fodro/saberchan/internal/ban Store

// PermanentTTL is the Redis TTL used for permanent bans.
const PermanentTTL = 100 * 365 * 24 * time.Hour

var (
	ErrReasonRequired = errors.New("reason is required")
	ErrNoTarget       = errors.New("at least one of ip or fingerprint is required")
	ErrPostNotFound   = errors.New("post not found")
	ErrAlreadyDeleted = errors.New("post already deleted")
	ErrNoIdentifier   = errors.New("post has no ip or fingerprint to ban")
)

// Record is the JSON value stored at a ban key.
type Record struct {
	Reason string    `json:"reason"`
	Until  time.Time `json:"until"`
}

// Store persists ban records keyed by "ban:ip:{ip}" or "ban:fp:{fp}".
type Store interface {
	SetBan(ctx context.Context, key string, record Record, ttl time.Duration) error
	// GetBan returns (nil, nil) if the key doesn't exist (i.e. not banned).
	GetBan(ctx context.Context, key string) (*Record, error)
}

// BanRequest describes a ban to place. At least one of IP/Fingerprint must
// be set. Duration <= 0 means permanent.
type BanRequest struct {
	Reason      string
	Duration    time.Duration
	IP          string
	Fingerprint string
}

type Service interface {
	// Check looks up ban:ip:{ip} then ban:fp:{fp}, returning the first hit.
	// Returns (nil, nil) when neither is banned.
	Check(ctx context.Context, ip, fingerprint string) (*Record, error)
	// Ban places a ban on the given IP and/or fingerprint.
	Ban(ctx context.Context, req BanRequest) error
	// BanPost bans the author of postID (preferring IP over fingerprint)
	// and soft-deletes the post.
	BanPost(ctx context.Context, postID uuid.UUID, reason string, duration time.Duration) error
}
