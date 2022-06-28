package services

import (
	"context"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
)

// UserService - бизнес-логика для работы с пользователями
type UserService struct {
	cfg  *config.Config
	repo repo.Repo
}

func NewUserService(cfg *config.Config, repo repo.Repo) *UserService {
	return &UserService{cfg: cfg, repo: repo}
}

// Create - создает нового пользователя
func (s UserService) Create(ctx context.Context, user *models.User) error {
	if err := s.repo.UserCreate(ctx, user); err != nil {
		return ErrInternal
	}
	return nil
}
