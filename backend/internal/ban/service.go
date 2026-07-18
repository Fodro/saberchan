package ban

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func ipKey(ip string) string          { return "ban:ip:" + ip }
func fingerprintKey(fp string) string { return "ban:fp:" + fp }

type service struct {
	store Store
	repo  database.Repository
}

// NewService builds a ban Service. repo is used by BanPost to look up the
// post's IP/fingerprint and to soft-delete it.
func NewService(store Store, repo database.Repository) Service {
	return &service{store: store, repo: repo}
}

func (s *service) Check(ctx context.Context, ip, fingerprint string) (*Record, error) {
	if ip != "" {
		rec, err := s.store.GetBan(ctx, ipKey(ip))
		if err != nil {
			return nil, err
		}
		if rec != nil {
			return rec, nil
		}
	}
	if fingerprint != "" {
		rec, err := s.store.GetBan(ctx, fingerprintKey(fingerprint))
		if err != nil {
			return nil, err
		}
		if rec != nil {
			return rec, nil
		}
	}
	return nil, nil
}

func (s *service) Ban(ctx context.Context, req BanRequest) error {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return ErrReasonRequired
	}
	if req.IP == "" && req.Fingerprint == "" {
		return ErrNoTarget
	}

	ttl := req.Duration
	var until time.Time
	if ttl > 0 {
		until = time.Now().Add(ttl)
	} else {
		ttl = PermanentTTL
		until = time.Now().Add(PermanentTTL)
	}

	record := Record{Reason: reason, Until: until}

	if req.IP != "" {
		if err := s.store.SetBan(ctx, ipKey(req.IP), record, ttl); err != nil {
			return err
		}
	}
	if req.Fingerprint != "" {
		if err := s.store.SetBan(ctx, fingerprintKey(req.Fingerprint), record, ttl); err != nil {
			return err
		}
	}
	return nil
}

// BanPost bans the post's IP if present, else its fingerprint, then
// soft-deletes the post. Per spec: prefer IP when non-empty, else
// fingerprint.
func (s *service) BanPost(ctx context.Context, postID uuid.UUID, reason string, duration time.Duration) error {
	post, err := s.repo.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPostNotFound
		}
		return err
	}
	if post.DeletedAt != nil {
		return ErrAlreadyDeleted
	}

	req := BanRequest{Reason: reason, Duration: duration}
	switch {
	case post.IP != "":
		req.IP = post.IP
	case post.BrowserFingerprint != "":
		req.Fingerprint = post.BrowserFingerprint
	default:
		return ErrNoIdentifier
	}

	if err := s.Ban(ctx, req); err != nil {
		return err
	}
	return s.repo.SoftDeletePost(ctx, postID)
}
