package usecases

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/ofstudio/go-shortener/internal/repo"
)

var errDBNotInitialized = errors.New("db is not initialized")

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
	if sqlRepo, ok := u.repo.(*repo.SQLRepo); ok {
		if db := sqlRepo.DB(); db != nil {
			if err := db.PingContext(ctx); err != nil {
				log.Err(err).Msg("failed to ping db")
				return err
			}
		} else {
			log.Err(errDBNotInitialized).Msg("failed to ping db")
			return errDBNotInitialized
		}
	}
	return nil
}
