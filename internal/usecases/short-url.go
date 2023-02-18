package usecases

import (
	"context"
	"errors"
	"net/url"

	"github.com/rs/zerolog/log"

	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/pkgerrors"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/pkg/shortid"
)

// ShortURL - бизнес-логика для сокращенных ссылок
type ShortURL struct {
	repo    repo.IRepo
	stopCtx context.Context // Контекст для остановки фоновых задач
	baseURL string
}

// NewShortURL - конструктор ShortURL
func NewShortURL(stopCtx context.Context, repo repo.IRepo, baseURL string) *ShortURL {
	return &ShortURL{
		stopCtx: stopCtx,
		repo:    repo,
		baseURL: baseURL,
	}
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
		return nil, pkgerrors.ErrNotFound
	} else if err != nil {
		log.Err(err).Msg("failed to get user by id")
		return nil, pkgerrors.ErrInternal
	}

	// Создаем модель и сохраняем в репозиторий
	shortURL := &models.ShortURL{
		ID:          shortid.Generate(),
		OriginalURL: OriginalURL,
		UserID:      userID,
	}
	err = u.repo.ShortURLCreate(ctx, shortURL)

	if err != nil && !errors.Is(err, repo.ErrDuplicate) {
		log.Err(err).Msg("failed to create short url")
		return nil, pkgerrors.ErrInternal
	}

	// Если такой URL уже существует,
	// запрашиваем его и возвращаем ErrDuplicate
	if errors.Is(err, repo.ErrDuplicate) {
		shortURL, err = u.GetByOriginalURL(ctx, OriginalURL)
		if err != nil {
			return nil, err
		}
		return shortURL, pkgerrors.ErrDuplicate
	}

	// Возвращаем модель
	return shortURL, nil
}

// GetByID - возвращает ShortURL по его id
func (u ShortURL) GetByID(ctx context.Context, id string) (*models.ShortURL, error) {
	shortURL, err := u.repo.ShortURLGetByID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, pkgerrors.ErrNotFound
	} else if err != nil {
		log.Err(err).Msg("failed to get short url by id")
		return nil, pkgerrors.ErrInternal
	}
	// Проверяем, не помечена ли ссылка как удаленная
	if shortURL.Deleted {
		return nil, pkgerrors.ErrDeleted
	}
	return shortURL, nil
}

// GetByUserID - возвращает все ShortURL пользователя
func (u ShortURL) GetByUserID(ctx context.Context, id uint) ([]models.ShortURL, error) {
	shortURLs, err := u.repo.ShortURLGetByUserID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, pkgerrors.ErrNotFound
	} else if err != nil {
		log.Err(err).Msg("failed to get short urls by user id")
		return nil, pkgerrors.ErrInternal
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
		return nil, pkgerrors.ErrNotFound
	} else if err != nil {
		log.Err(err).Msg("failed to get short url by original url")
		return nil, pkgerrors.ErrInternal
	}
	// Проверяем, не помечена ли ссылка как удаленная
	if shortURL.Deleted {
		return nil, pkgerrors.ErrDeleted
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

	// Для дальнейших вызовов в качестве контекста используем контекст для фоновых задач.
	// В противном случае, если контекст запроса будет отменен, то все вызовы будут прерваны
	chans := fanOut(u.stopCtx, ch)
	if _, err := u.repo.ShortURLDeleteBatch(u.stopCtx, userID, chans...); err != nil {
		log.Err(err).Msg("failed to batch delete short urls")
		return pkgerrors.ErrInternal
	}
	return nil
}

// Count - возвращает количество сокращенных ссылок
func (u ShortURL) Count(ctx context.Context) (int, error) {
	count, err := u.repo.ShortURLCount(ctx)
	if err != nil {
		log.Err(err).Msg("failed to count short urls")
		return 0, pkgerrors.ErrInternal
	}
	return count, nil
}

// Resolve - возвращает сокращенный URL по его id
func (u ShortURL) Resolve(id string) string {
	return u.baseURL + id
}

// validateURL - проверяет URL на максимальную длину и http/https-протокол
func (u ShortURL) validateURL(rawURL string) error {
	// Проверка на максимальную длину URL
	if len(rawURL) > models.URLMaxLen {
		return pkgerrors.ErrValidation
	}

	// Проверка на валидный URL
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return pkgerrors.ErrValidation
	}

	// Проверка на http / https
	// NB: URL.Scheme всегда будет в нижнем регистре (специально приводить не надо)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return pkgerrors.ErrValidation
	}

	return nil
}
