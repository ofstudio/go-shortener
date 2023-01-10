package services

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/repo"
)

// Container - контейнер сервисов
type Container struct {
	ShortURLService *ShortURLService
	HealthService   *HealthService
	UserService     *UserService
}

// NewContainer - конструктор Container
func NewContainer(cfg *config.Config, repo repo.Repo) *Container {
	return &Container{
		ShortURLService: NewShortURLService(cfg, repo),
		HealthService:   NewHealthService(repo),
		UserService:     NewUserService(cfg, repo),
	}
}
