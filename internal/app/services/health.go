package services

import (
	"context"

	"github.com/ofstudio/go-shortener/internal/repo"
)

// HealthService - healthcheck-сервис приложения
type HealthService struct {
	repo repo.Repo
}

// NewHealthService - конструктор HealthService
func NewHealthService(repo repo.Repo) *HealthService {
	return &HealthService{repo: repo}
}

// Check - выполняет проверку приложения
func (s *HealthService) Check(ctx context.Context) error {
	// Если используется SQL-репозиторий, то проверяем подключение к БД.
	sqlRepo, ok := s.repo.(*repo.SQLRepo)
	if ok {
		db := sqlRepo.DB()
		if db != nil {
			return db.PingContext(ctx)
		}
	}
	return nil
}
