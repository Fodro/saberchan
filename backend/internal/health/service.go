package health

import "github.com/Fodro/saberchan/internal/database"

type service struct {
	repo database.Repository
}

func (s *service) Readiness() error {
	return s.repo.Ping()
}

func NewService(repo database.Repository) Service {
	return &service{repo: repo}
}
