package usecases

import (
	"context"

	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/repo"
)

// Container - контейнер сервисов
type Container struct {
	ShortURL *ShortURL
	User     *User
	Health   *Health
}

// NewContainer - конструктор Container
func NewContainer(ctx context.Context, cfg *config.Config, repo repo.IRepo) *Container {
	return &Container{
		ShortURL: NewShortURL(ctx, repo, cfg.BaseURL.String()),
		User:     NewUser(repo),
		Health:   NewHealth(repo),
	}
}
