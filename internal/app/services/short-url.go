package services

import (
	"context"
	"errors"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/pkg/shortid"
	"net/url"
)

// ShortURLService - бизнес-логика для сокращенных ссылок
type ShortURLService struct {
	cfg  *config.Config
	repo repo.Repo
}

func NewShortURLService(cfg *config.Config, repo repo.Repo) *ShortURLService {
	return &ShortURLService{cfg, repo}
}

// Create - создает и возвращает ShortURL
func (s ShortURLService) Create(ctx context.Context, userID uint, OriginalURL string) (*models.ShortURL, error) {
	// Проверяем URL на валидность
	if err := s.validateURL(OriginalURL); err != nil {
		return nil, err
	}

	// Проверяем, существует ли такой пользователь
	_, err := s.repo.UserGetByID(ctx, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}

	// Создаем модель и сохраняем в репозиторий
	shortURL := &models.ShortURL{
		ID:          shortid.Generate(),
		OriginalURL: OriginalURL,
		UserID:      userID,
	}
	err = s.repo.ShortURLCreate(ctx, shortURL)

	// Если такой URL уже существует, возвращаем ErrDuplicate
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrDuplicate
	} else if err != nil {
		return nil, ErrInternal
	}
	// Возвращаем модель
	return shortURL, nil
}

// GetByID - возвращает ShortURL по его id
func (s ShortURLService) GetByID(ctx context.Context, id string) (*models.ShortURL, error) {
	shortURL, err := s.repo.ShortURLGetByID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}
	return shortURL, nil
}

// GetByUserID - возвращает все ShortURL пользователя
func (s ShortURLService) GetByUserID(ctx context.Context, id uint) ([]models.ShortURL, error) {
	shortURLs, err := s.repo.ShortURLGetByUserID(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}
	return shortURLs, nil
}

// GetByOriginalURL - возвращает ShortURL по его оригинальному URL
func (s ShortURLService) GetByOriginalURL(ctx context.Context, rawURL string) (*models.ShortURL, error) {
	shortURL, err := s.repo.ShortURLGetByOriginalURL(ctx, rawURL)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}
	return shortURL, nil
}

func (s ShortURLService) Resolve(id string) string {
	return s.cfg.BaseURL.String() + id
}

// validateURL - проверяет URL на максимальную длину и http/https-протокол
func (s ShortURLService) validateURL(rawURL string) error {
	// Проверка на максимальную длину URL
	if len(rawURL) > models.URLMaxLen {
		return ErrValidation
	}

	// Проверка на валидный URL
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ErrValidation
	}

	// Проверка на http / https
	// NB: URL.Scheme всегда будет в нижнем регистре (специально приводить не надо)
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrValidation
	}

	return nil
}
