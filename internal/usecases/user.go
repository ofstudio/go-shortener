package usecases

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/pkgerrors"
	"github.com/ofstudio/go-shortener/internal/repo"
)

// User - бизнес-логика для работы с пользователями
type User struct {
	repo repo.IRepo
}

// NewUser - конструктор User
func NewUser(repo repo.IRepo) *User {
	return &User{repo: repo}
}

// Create - создает нового пользователя
func (u User) Create(ctx context.Context, user *models.User) error {
	if err := u.repo.UserCreate(ctx, user); err != nil {
		log.Err(err).Msg("failed to create user")
		return pkgerrors.ErrInternal
	}
	return nil
}

// Count - возвращает количество пользователей
func (u User) Count(ctx context.Context) (int, error) {
	count, err := u.repo.UserCount(ctx)
	if err != nil {
		log.Err(err).Msg("failed to count users")
		return 0, pkgerrors.ErrInternal
	}
	return count, nil
}
