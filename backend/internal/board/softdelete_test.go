package board

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	dbmocks "github.com/Fodro/saberchan/internal/database/mocks"
	filemocks "github.com/Fodro/saberchan/internal/file/mocks"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func newTestService(ctrl *gomock.Controller) (*dbmocks.MockRepository, Service) {
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	return repo, NewService(repo, files, &config.Config{}, nil)
}

func TestDeleteBoard_SoftDeletesWhenLive(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	repo.EXPECT().GetBoardById(gomock.Any(), id, true).Return(&database.Board{ID: id}, nil)
	repo.EXPECT().SoftDeleteBoard(gomock.Any(), id).Return(nil)

	if err := svc.DeleteBoard(context.Background(), id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDeleteBoard_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	now := time.Now()
	repo.EXPECT().GetBoardById(gomock.Any(), id, true).Return(&database.Board{ID: id, DeletedAt: &now}, nil)

	err := svc.DeleteBoard(context.Background(), id)
	if !errors.Is(err, ErrAlreadyDeleted) {
		t.Fatalf("got %v, want ErrAlreadyDeleted", err)
	}
}

func TestDeleteBoard_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	repo.EXPECT().GetBoardById(gomock.Any(), id, true).Return(nil, pgx.ErrNoRows)

	err := svc.DeleteBoard(context.Background(), id)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestRestoreBoard_WithinWindow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	deletedAt := time.Now().Add(-1 * time.Hour)
	repo.EXPECT().GetBoardById(gomock.Any(), id, true).Return(&database.Board{ID: id, DeletedAt: &deletedAt}, nil)
	repo.EXPECT().RestoreBoard(gomock.Any(), id).Return(nil)

	if err := svc.RestoreBoard(context.Background(), id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRestoreBoard_ExpiredWindow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	deletedAt := time.Now().Add(-25 * time.Hour)
	repo.EXPECT().GetBoardById(gomock.Any(), id, true).Return(&database.Board{ID: id, DeletedAt: &deletedAt}, nil)

	err := svc.RestoreBoard(context.Background(), id)
	if !errors.Is(err, ErrRestoreExpired) {
		t.Fatalf("got %v, want ErrRestoreExpired", err)
	}
}

func TestRestoreBoard_NotCurrentlyDeleted(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	repo.EXPECT().GetBoardById(gomock.Any(), id, true).Return(&database.Board{ID: id}, nil)

	err := svc.RestoreBoard(context.Background(), id)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestDeleteThread_SoftDeletesWhenLive(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	repo.EXPECT().GetThread(gomock.Any(), id, true).Return(&database.Thread{ID: id}, nil)
	repo.EXPECT().SoftDeleteThread(gomock.Any(), id).Return(nil)

	if err := svc.DeleteThread(context.Background(), id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRestoreThread_ExpiredWindow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	deletedAt := time.Now().Add(-25 * time.Hour)
	repo.EXPECT().GetThread(gomock.Any(), id, true).Return(&database.Thread{ID: id, DeletedAt: &deletedAt}, nil)

	err := svc.RestoreThread(context.Background(), id)
	if !errors.Is(err, ErrRestoreExpired) {
		t.Fatalf("got %v, want ErrRestoreExpired", err)
	}
}

func TestDeletePost_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	now := time.Now()
	repo.EXPECT().GetPost(gomock.Any(), id).Return(&database.Post{ID: id, DeletedAt: &now}, nil)

	err := svc.DeletePost(context.Background(), id)
	if !errors.Is(err, ErrAlreadyDeleted) {
		t.Fatalf("got %v, want ErrAlreadyDeleted", err)
	}
}

func TestRestorePost_WithinWindow(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	deletedAt := time.Now().Add(-1 * time.Hour)
	repo.EXPECT().GetPost(gomock.Any(), id).Return(&database.Post{ID: id, DeletedAt: &deletedAt}, nil)
	repo.EXPECT().RestorePost(gomock.Any(), id).Return(nil)

	if err := svc.RestorePost(context.Background(), id); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRestorePost_AlreadyPurged(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo, svc := newTestService(ctrl)

	id := uuid.New()
	deletedAt := time.Now().Add(-1 * time.Hour)
	purgedAt := time.Now()
	repo.EXPECT().GetPost(gomock.Any(), id).Return(&database.Post{ID: id, DeletedAt: &deletedAt, PurgedAt: &purgedAt}, nil)

	err := svc.RestorePost(context.Background(), id)
	if !errors.Is(err, ErrRestoreExpired) {
		t.Fatalf("got %v, want ErrRestoreExpired", err)
	}
}
