package health

import (
	"errors"
	"testing"

	"github.com/Fodro/saberchan/internal/database/mocks"
	"go.uber.org/mock/gomock"
)

func TestReadiness(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)

	repo.EXPECT().Ping().Return(nil)
	if err := NewService(repo).Readiness(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	repo.EXPECT().Ping().Return(errors.New("down"))
	if err := NewService(repo).Readiness(); err == nil {
		t.Fatal("expected error")
	}
}
