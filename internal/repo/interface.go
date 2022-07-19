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
	// ShortURLDelete - помечает удаленной короткую ссылку пользователя по ее id.
	ShortURLDelete(context.Context, uint, string) error
	// ShortURLDeleteBatch - помечает удаленными несколько сокращенных ссылок пользователя по их id.
	// Принимает на вход список каналов для передачи идентификаторов.
	// Возвращает количество удаленных сокращенных ссылок.
	ShortURLDeleteBatch(context.Context, uint, ...chan string) (int64, error)
	Close() error
}
