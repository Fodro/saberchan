package ban_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fodro/saberchan/internal/ban"
	"github.com/Fodro/saberchan/internal/ban/mocks"
	"github.com/Fodro/saberchan/internal/database"
	dbmocks "github.com/Fodro/saberchan/internal/database/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func newTestService(ctrl *gomock.Controller) (*mocks.MockStore, *dbmocks.MockRepository, ban.Service) {
	store := mocks.NewMockStore(ctrl)
	repo := dbmocks.NewMockRepository(ctrl)
	return store, repo, ban.NewService(store, repo)
}

func TestCheck_PrefersIPOverFingerprint(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	want := &ban.Record{Reason: "spam", Until: time.Now().Add(time.Hour)}
	store.EXPECT().GetBan(ctx, "ban:ip:1.2.3.4").Return(want, nil)

	got, err := svc.Check(ctx, "1.2.3.4", "fp123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCheck_FallsBackToFingerprint(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	want := &ban.Record{Reason: "spam", Until: time.Now().Add(time.Hour)}
	store.EXPECT().GetBan(ctx, "ban:ip:1.2.3.4").Return(nil, nil)
	store.EXPECT().GetBan(ctx, "ban:fp:fp123").Return(want, nil)

	got, err := svc.Check(ctx, "1.2.3.4", "fp123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCheck_NoBanReturnsNil(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	store.EXPECT().GetBan(ctx, "ban:ip:1.2.3.4").Return(nil, nil)
	store.EXPECT().GetBan(ctx, "ban:fp:fp123").Return(nil, nil)

	got, err := svc.Check(ctx, "1.2.3.4", "fp123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestCheck_SkipsEmptyIdentifiers(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	// No IP, so only fp lookup should happen.
	store.EXPECT().GetBan(ctx, "ban:fp:fp123").Return(nil, nil)

	got, err := svc.Check(ctx, "", "fp123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestBan_RequiresReason(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, _, svc := newTestService(ctrl)

	err := svc.Ban(context.Background(), ban.BanRequest{IP: "1.2.3.4"})
	if !errors.Is(err, ban.ErrReasonRequired) {
		t.Fatalf("got %v, want ErrReasonRequired", err)
	}
}

func TestBan_RequiresTarget(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, _, svc := newTestService(ctrl)

	err := svc.Ban(context.Background(), ban.BanRequest{Reason: "spam"})
	if !errors.Is(err, ban.ErrNoTarget) {
		t.Fatalf("got %v, want ErrNoTarget", err)
	}
}

func TestBan_TemporaryUsesDurationTTL(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	var gotTTL time.Duration
	var gotRecord ban.Record
	store.EXPECT().SetBan(ctx, "ban:ip:1.2.3.4", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, record ban.Record, ttl time.Duration) error {
			gotRecord = record
			gotTTL = ttl
			return nil
		})

	if err := svc.Ban(ctx, ban.BanRequest{Reason: "spam", Duration: time.Hour, IP: "1.2.3.4"}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if gotTTL != time.Hour {
		t.Fatalf("got ttl %v, want 1h", gotTTL)
	}
	if gotRecord.Reason != "spam" {
		t.Fatalf("got reason %q, want spam", gotRecord.Reason)
	}
	wantUntil := time.Now().Add(time.Hour)
	if gotRecord.Until.Sub(wantUntil).Abs() > 5*time.Second {
		t.Fatalf("got until %v, want ~%v", gotRecord.Until, wantUntil)
	}
}

func TestBan_PermanentUsesPermanentTTL(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	var gotTTL time.Duration
	store.EXPECT().SetBan(ctx, "ban:fp:fp123", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ ban.Record, ttl time.Duration) error {
			gotTTL = ttl
			return nil
		})

	if err := svc.Ban(ctx, ban.BanRequest{Reason: "spam", Duration: 0, Fingerprint: "fp123"}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if gotTTL != ban.PermanentTTL {
		t.Fatalf("got ttl %v, want %v", gotTTL, ban.PermanentTTL)
	}
}

func TestBan_SetsBothWhenBothProvided(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, _, svc := newTestService(ctrl)
	ctx := context.Background()

	store.EXPECT().SetBan(ctx, "ban:ip:1.2.3.4", gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().SetBan(ctx, "ban:fp:fp123", gomock.Any(), gomock.Any()).Return(nil)

	err := svc.Ban(ctx, ban.BanRequest{Reason: "spam", Duration: time.Hour, IP: "1.2.3.4", Fingerprint: "fp123"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBanPost_PrefersIP(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, repo, svc := newTestService(ctrl)
	ctx := context.Background()

	postID := uuid.New()
	repo.EXPECT().GetPost(ctx, postID).Return(&database.Post{
		ID: postID, IP: "1.2.3.4", BrowserFingerprint: "fp123",
	}, nil)
	store.EXPECT().SetBan(ctx, "ban:ip:1.2.3.4", gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().SoftDeletePost(ctx, postID).Return(nil)

	if err := svc.BanPost(ctx, postID, "spam", time.Hour); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBanPost_FallsBackToFingerprintWhenIPEmpty(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store, repo, svc := newTestService(ctrl)
	ctx := context.Background()

	postID := uuid.New()
	repo.EXPECT().GetPost(ctx, postID).Return(&database.Post{
		ID: postID, IP: "", BrowserFingerprint: "fp123",
	}, nil)
	store.EXPECT().SetBan(ctx, "ban:fp:fp123", gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().SoftDeletePost(ctx, postID).Return(nil)

	if err := svc.BanPost(ctx, postID, "spam", time.Hour); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBanPost_NoIdentifier(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, repo, svc := newTestService(ctrl)
	ctx := context.Background()

	postID := uuid.New()
	repo.EXPECT().GetPost(ctx, postID).Return(&database.Post{ID: postID}, nil)

	err := svc.BanPost(ctx, postID, "spam", time.Hour)
	if !errors.Is(err, ban.ErrNoIdentifier) {
		t.Fatalf("got %v, want ErrNoIdentifier", err)
	}
}

func TestBanPost_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, repo, svc := newTestService(ctrl)
	ctx := context.Background()

	postID := uuid.New()
	repo.EXPECT().GetPost(ctx, postID).Return(nil, pgx.ErrNoRows)

	err := svc.BanPost(ctx, postID, "spam", time.Hour)
	if !errors.Is(err, ban.ErrPostNotFound) {
		t.Fatalf("got %v, want ErrPostNotFound", err)
	}
}

func TestBanPost_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, repo, svc := newTestService(ctrl)
	ctx := context.Background()

	postID := uuid.New()
	now := time.Now()
	repo.EXPECT().GetPost(ctx, postID).Return(&database.Post{ID: postID, IP: "1.2.3.4", DeletedAt: &now}, nil)

	err := svc.BanPost(ctx, postID, "spam", time.Hour)
	if !errors.Is(err, ban.ErrAlreadyDeleted) {
		t.Fatalf("got %v, want ErrAlreadyDeleted", err)
	}
}

func TestBanPost_EmptyReasonRejected(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	_, repo, svc := newTestService(ctrl)
	ctx := context.Background()

	postID := uuid.New()
	repo.EXPECT().GetPost(ctx, postID).Return(&database.Post{ID: postID, IP: "1.2.3.4"}, nil)

	err := svc.BanPost(ctx, postID, "  ", time.Hour)
	if !errors.Is(err, ban.ErrReasonRequired) {
		t.Fatalf("got %v, want ErrReasonRequired", err)
	}
}
