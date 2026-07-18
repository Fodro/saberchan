package follow

import (
	"context"

	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

type service struct {
	store Store
	repo  database.Repository
}

// NewService builds a follow Service.
func NewService(store Store, repo database.Repository) Service {
	return &service{store: store, repo: repo}
}

func (s *service) Follow(ctx context.Context, threadID uuid.UUID) error {
	infos, err := s.repo.GetFollowThreadInfos(ctx, []uuid.UUID{threadID})
	if err != nil {
		return err
	}
	if len(infos) == 0 || !infos[0].BelowBumpLimit {
		return ErrDead
	}
	return s.store.Touch(ctx, ThreadKey(threadID), TTL)
}

func (s *service) RefreshOnBump(ctx context.Context, threadID uuid.UUID) error {
	exists, err := s.store.Exists(ctx, ThreadKey(threadID))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	infos, err := s.repo.GetFollowThreadInfos(ctx, []uuid.UUID{threadID})
	if err != nil {
		return err
	}
	if len(infos) == 0 || !infos[0].BelowBumpLimit {
		return nil
	}
	return s.store.Touch(ctx, ThreadKey(threadID), TTL)
}

func (s *service) Status(ctx context.Context, ids []uuid.UUID) ([]Status, error) {
	if len(ids) == 0 {
		return []Status{}, nil
	}

	infos, err := s.repo.GetFollowThreadInfos(ctx, ids)
	if err != nil {
		return nil, err
	}
	infoByID := make(map[uuid.UUID]database.FollowThreadInfo, len(infos))
	for _, info := range infos {
		infoByID[info.ID] = info
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = ThreadKey(id)
	}
	exists, err := s.store.ExistsMany(ctx, keys)
	if err != nil {
		return nil, err
	}

	out := make([]Status, len(ids))
	for i, id := range ids {
		st := Status{ID: id, Dead: true}
		info, ok := infoByID[id]
		if ok {
			st.Title = info.Title
			st.BoardAlias = info.BoardAlias
			st.RepliesCount = info.RepliesCount
			st.Dead = !exists[i] || !info.BelowBumpLimit
		}
		out[i] = st
	}
	return out, nil
}
