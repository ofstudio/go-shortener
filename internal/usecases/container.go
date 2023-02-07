package usecases

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/repo"
)

// Container - контейнер сервисов
type Container struct {
	ShortURL *ShortURL
	User     *User
	Health   *Health
}

// NewContainer - конструктор Container
func NewContainer(cfg *config.Config, repo repo.IRepo) *Container {
	return &Container{
		ShortURL: NewShortURL(cfg, repo),
		User:     NewUser(cfg, repo),
		Health:   NewHealth(repo),
	}
}
