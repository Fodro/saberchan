package follow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fodro/saberchan/internal/database"
	dbmocks "github.com/Fodro/saberchan/internal/database/mocks"
	"github.com/Fodro/saberchan/internal/follow"
	"github.com/Fodro/saberchan/internal/follow/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func newTestService(ctrl *gomock.Controller) (*mocks.MockStore, *dbmocks.MockRepository, follow.Service) {
	store := mocks.NewMockStore(ctrl)
	repo := dbmocks.NewMockRepository(ctrl)
	return store, repo, follow.NewService(store, repo)
}

func TestFollow_TouchesWhenBelowBumpLimit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, repo, svc := newTestService(ctrl)
	ctx := context.Background()
	id := uuid.New()

	repo.EXPECT().GetFollowThreadInfos(ctx, []uuid.UUID{id}).Return([]database.FollowThreadInfo{{
		ID: id, Title: "hi", BoardAlias: "b", RepliesCount: 3, BelowBumpLimit: true,
	}}, nil)
	store.EXPECT().Touch(ctx, follow.ThreadKey(id), follow.TTL).Return(nil)

	if err := svc.Follow(ctx, id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestFollow_DeadWhenBumpLimited(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, repo, svc := newTestService(ctrl)
	ctx := context.Background()
	id := uuid.New()

	repo.EXPECT().GetFollowThreadInfos(ctx, []uuid.UUID{id}).Return([]database.FollowThreadInfo{{
		ID: id, BelowBumpLimit: false,
	}}, nil)

	err := svc.Follow(ctx, id)
	if !errors.Is(err, follow.ErrDead) {
		t.Fatalf("got %v, want ErrDead", err)
	}
}

func TestFollow_DeadWhenMissing(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, repo, svc := newTestService(ctrl)
	ctx := context.Background()
	id := uuid.New()

	repo.EXPECT().GetFollowThreadInfos(ctx, []uuid.UUID{id}).Return(nil, nil)

	err := svc.Follow(ctx, id)
	if !errors.Is(err, follow.ErrDead) {
		t.Fatalf("got %v, want ErrDead", err)
	}
}

func TestRefreshOnBump_RefreshesWhenKeyExistsAndAlive(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, repo, svc := newTestService(ctrl)
	ctx := context.Background()
	id := uuid.New()

	store.EXPECT().Exists(ctx, follow.ThreadKey(id)).Return(true, nil)
	repo.EXPECT().GetFollowThreadInfos(ctx, []uuid.UUID{id}).Return([]database.FollowThreadInfo{{
		ID: id, BelowBumpLimit: true,
	}}, nil)
	store.EXPECT().Touch(ctx, follow.ThreadKey(id), follow.TTL).Return(nil)

	if err := svc.RefreshOnBump(ctx, id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRefreshOnBump_NoopWhenKeyMissing(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()
	id := uuid.New()

	store.EXPECT().Exists(ctx, follow.ThreadKey(id)).Return(false, nil)

	if err := svc.RefreshOnBump(ctx, id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRefreshOnBump_NoopWhenBumpLimited(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, repo, svc := newTestService(ctrl)
	ctx := context.Background()
	id := uuid.New()

	store.EXPECT().Exists(ctx, follow.ThreadKey(id)).Return(true, nil)
	repo.EXPECT().GetFollowThreadInfos(ctx, []uuid.UUID{id}).Return([]database.FollowThreadInfo{{
		ID: id, BelowBumpLimit: false,
	}}, nil)

	if err := svc.RefreshOnBump(ctx, id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestStatus_MarksDeadByKeyOrBumpLimit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, repo, svc := newTestService(ctrl)
	ctx := context.Background()
	aliveID := uuid.New()
	noKeyID := uuid.New()
	limitedID := uuid.New()
	missingID := uuid.New()
	ids := []uuid.UUID{aliveID, noKeyID, limitedID, missingID}

	repo.EXPECT().GetFollowThreadInfos(ctx, ids).Return([]database.FollowThreadInfo{
		{ID: aliveID, Title: "a", BoardAlias: "b", RepliesCount: 1, BelowBumpLimit: true},
		{ID: noKeyID, Title: "n", BoardAlias: "b", RepliesCount: 2, BelowBumpLimit: true},
		{ID: limitedID, Title: "l", BoardAlias: "b", RepliesCount: 499, BelowBumpLimit: false},
	}, nil)
	store.EXPECT().ExistsMany(ctx, []string{
		follow.ThreadKey(aliveID),
		follow.ThreadKey(noKeyID),
		follow.ThreadKey(limitedID),
		follow.ThreadKey(missingID),
	}).Return([]bool{true, false, true, false}, nil)

	got, err := svc.Status(ctx, ids)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Dead || got[0].Title != "a" || got[0].RepliesCount != 1 {
		t.Fatalf("alive status: %+v", got[0])
	}
	if !got[1].Dead || got[1].Title != "n" {
		t.Fatalf("no-key status: %+v", got[1])
	}
	if !got[2].Dead || got[2].Title != "l" {
		t.Fatalf("limited status: %+v", got[2])
	}
	if !got[3].Dead || got[3].Title != "" {
		t.Fatalf("missing status: %+v", got[3])
	}
}

func TestStatus_EmptyIDs(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, _, svc := newTestService(ctrl)

	got, err := svc.Status(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestTTL(t *testing.T) {
	t.Parallel()
	if follow.TTL != 7*24*time.Hour {
		t.Fatalf("TTL=%v", follow.TTL)
	}
}
