package health

import (
	"context"

	"github.com/Fodro/saberchan/internal/database"
)

type service struct {
	repo database.Repository
}

func (s *service) Readiness(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func NewService(repo database.Repository) Service {
	return &service{repo: repo}
}
