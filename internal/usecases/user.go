package usecases

import (
	"context"

	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
)

// User - бизнес-логика для работы с пользователями
type User struct {
	cfg  *config.Config
	repo repo.IRepo
}

// NewUser - конструктор User
func NewUser(cfg *config.Config, repo repo.IRepo) *User {
	return &User{cfg: cfg, repo: repo}
}

// Create - создает нового пользователя
func (u User) Create(ctx context.Context, user *models.User) error {
	if err := u.repo.UserCreate(ctx, user); err != nil {
		return app.ErrInternal
	}
	return nil
}

// Count - возвращает количество пользователей
func (u User) Count(ctx context.Context) (int, error) {
	count, err := u.repo.UserCount(ctx)
	if err != nil {
		return 0, app.ErrInternal
	}
	return count, nil
}
