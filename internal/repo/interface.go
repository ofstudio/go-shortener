package repo

import (
	"context"
	"github.com/ofstudio/go-shortener/internal/models"
)

// Repo - интерфейс репозитория.
type Repo interface {
	// UserCreate - добавляет нового пользователя в репозиторий.
	UserCreate(context.Context, *models.User) error
	// UserGetByID - возвращает пользователя по его id.
	UserGetByID(context.Context, uint) (*models.User, error)
	// ShortURLCreate - добавляет новую сокращенную ссылку в репозиторий.
	ShortURLCreate(context.Context, *models.ShortURL) error
	// ShortURLGetByID - возвращает сокращенную ссылку по ее id.
	ShortURLGetByID(context.Context, string) (*models.ShortURL, error)
	// ShortURLGetByUserID - возвращает сокращенные ссылки пользователя.
	// Если пользователь не найден, или у пользователя нет ссылок возвращает nil.
	ShortURLGetByUserID(context.Context, uint) ([]models.ShortURL, error)
	// ShortURLGetByOriginalURL - возвращает сокращенную ссылку по ее оригинальному url.
	ShortURLGetByOriginalURL(context.Context, string) (*models.ShortURL, error)
	// ShortURLDeleteBatch - удаляет несколько сокращенных ссылок пользователя по их id.
	ShortURLDeleteBatch(context.Context, uint, []string) (int64, error)
	Close() error
}
