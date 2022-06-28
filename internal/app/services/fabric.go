package services

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/repo"
)

type Services struct {
	ShortURLService *ShortURLService
	HealthService   *HealthService
	UserService     *UserService
}

// Fabric - фабрика для создания сервисов.
func Fabric(cfg *config.Config, repo repo.Repo) *Services {
	return &Services{
		ShortURLService: NewShortURLService(cfg, repo),
		HealthService:   NewHealthService(repo),
		UserService:     NewUserService(cfg, repo),
	}
}
