package usecases

import (
	"context"
	"errors"
	"net/url"

	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/pkg/shortid"
)

// ShortURL - бизнес-логика для сокращенных ссылок
type ShortURL struct {
	cfg  *config.Config
	repo repo.IRepo
}

// NewShortURL - конструктор ShortURL
func NewShortURL(cfg *config.Config, repo repo.IRepo) *ShortURL {
	return &ShortURL{cfg, repo}
}

// Create - создает и возвращает ShortURL
func (u ShortURL) Create(ctx context.Context, userID uint, OriginalURL string) (*models.ShortURL, error) {
	// Проверяем URL на валидность
	if err := u.validateURL(OriginalURL); err != nil {
		return nil, err
	}

	// Проверяем, существует ли такой пользователь
	_, err := u.repo.UserGetByID(ctx, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, app.ErrNotFound
	} else if err != nil {
		return nil, app.ErrInternal
	}

	// Создаем модель и сохраняем в репозиторий
	shortURL := &models.ShortURL{
		ID:          shortid.Generate(),
		OriginalURL: OriginalURL,
		UserID:      userID,
	}
	err = u.repo.ShortURLCreate(ctx, shortURL)

	// Если такой URL уже существует, возвращаем ErrDuplicate
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, app.ErrDuplicate
	} else if err != nil {
		return nil, app.ErrInternal
	}
	// Возвращаем модель
	return shortURL, nil
}

// GetByID - возвращает ShortURL по его id
func (u ShortURL) GetByID(ctx context.Context, id string) (*models.ShortURL, error) {
	shortURL, err := u.repo.ShortURLGetByID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, app.ErrNotFound
	} else if err != nil {
		return nil, app.ErrInternal
	}
	// Проверяем, не помечена ли ссылка как удаленная
	if shortURL.Deleted {
		return nil, app.ErrDeleted
	}
	return shortURL, nil
}

// GetByUserID - возвращает все ShortURL пользователя
func (u ShortURL) GetByUserID(ctx context.Context, id uint) ([]models.ShortURL, error) {
	shortURLs, err := u.repo.ShortURLGetByUserID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, app.ErrNotFound
	} else if err != nil {
		return nil, app.ErrInternal
	}
	// Отфильтровываем ссылки, помеченные как удаленные
	var result []models.ShortURL
	for _, shortURL := range shortURLs {
		if !shortURL.Deleted {
			result = append(result, shortURL)
		}
	}
	return result, nil
}

// GetByOriginalURL - возвращает ShortURL по его оригинальному URL
func (u ShortURL) GetByOriginalURL(ctx context.Context, rawURL string) (*models.ShortURL, error) {
	shortURL, err := u.repo.ShortURLGetByOriginalURL(ctx, rawURL)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, app.ErrNotFound
	} else if err != nil {
		return nil, app.ErrInternal
	}
	// Проверяем, не помечена ли ссылка как удаленная
	if shortURL.Deleted {
		return nil, app.ErrDeleted
	}
	return shortURL, nil
}

// DeleteBatch - помечает удаленными несколько сокращенных ссылок пользователя по их id.
// Принимает на вход канал идентификаторов для удаления
func (u ShortURL) DeleteBatch(ctx context.Context, userID uint, ids []string) error {
	// Демультиплексируем слайс идентификаторов на каналы
	ch := make(chan string)
	go func() {
		for _, id := range ids {
			ch <- id
		}
		close(ch)
	}()

	chans := fanOut(ctx, ch)
	if _, err := u.repo.ShortURLDeleteBatch(ctx, userID, chans...); err != nil {
		return app.ErrInternal
	}
	return nil
}

// Count - возвращает количество сокращенных ссылок
func (u ShortURL) Count(ctx context.Context) (int, error) {
	count, err := u.repo.ShortURLCount(ctx)
	if err != nil {
		return 0, app.ErrInternal
	}
	return count, nil
}

// Resolve - возвращает сокращенный URL по его id
func (u ShortURL) Resolve(id string) string {
	return u.cfg.BaseURL.String() + id
}

// validateURL - проверяет URL на максимальную длину и http/https-протокол
func (u ShortURL) validateURL(rawURL string) error {
	// Проверка на максимальную длину URL
	if len(rawURL) > models.URLMaxLen {
		return app.ErrValidation
	}

	// Проверка на валидный URL
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return app.ErrValidation
	}

	// Проверка на http / https
	// NB: URL.Scheme всегда будет в нижнем регистре (специально приводить не надо)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return app.ErrValidation
	}

	return nil
}
