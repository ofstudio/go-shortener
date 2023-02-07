package usecases

import (
	"context"

	"github.com/ofstudio/go-shortener/internal/repo"
)

// Health - healthcheck-сервис приложения
type Health struct {
	repo repo.IRepo
}

// NewHealth - конструктор Health
func NewHealth(repo repo.IRepo) *Health {
	return &Health{repo: repo}
}

// Check - выполняет проверку приложения
func (u *Health) Check(ctx context.Context) error {
	// Если используется SQL-репозиторий, то проверяем подключение к БД.
	sqlRepo, ok := u.repo.(*repo.SQLRepo)
	if ok {
		db := sqlRepo.DB()
		if db != nil {
			return db.PingContext(ctx)
		}
	}
	return nil
}
