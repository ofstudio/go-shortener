package services

import (
	"github.com/ofstudio/go-shortener/internal/repo"
)

// HealthService - healthcheck-сервис приложения
type HealthService struct {
	repo repo.Repo
}

func NewHealthService(repo repo.Repo) *HealthService {
	return &HealthService{repo: repo}
}

// Check - выполняет проверку приложения
func (s *HealthService) Check() error {
	// Если используется SQL-репозиторий, то проверяем подключение к БД.
	sqlRepo, ok := s.repo.(*repo.SQLRepo)
	if ok {
		db := sqlRepo.DB()
		if db != nil {
			return db.Ping()
		}
	}
	return nil
}
