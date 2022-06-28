package repo

import "github.com/ofstudio/go-shortener/internal/models"

// Repo - интерфейс репозитория.
type Repo interface {
	// UserCreate - добавляет нового пользователя в репозиторий.
	UserCreate(*models.User) error
	// UserGetByID - возвращает пользователя по его id.
	UserGetByID(uint) (*models.User, error)
	// ShortURLCreate - добавляет новую сокращенную ссылку в репозиторий.
	ShortURLCreate(*models.ShortURL) error
	// ShortURLGetByID - возвращает сокращенную ссылку по ее id.
	ShortURLGetByID(string) (*models.ShortURL, error)
	// ShortURLGetByUserID - возвращает сокращенные ссылки пользователя.
	ShortURLGetByUserID(uint) ([]models.ShortURL, error)
	Close() error
}
