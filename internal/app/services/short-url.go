package services

import (
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
func (s ShortURLService) Create(userID uint, OriginalURL string) (*models.ShortURL, error) {
	// Проверяем, существует ли такой пользователь
	_, err := s.repo.UserGetByID(userID)
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}

	// Проверяем URL на валидность
	if err = s.validateURL(OriginalURL); err != nil {
		return nil, err
	}

	// Создаем модель и сохраняем в репозиторий
	shortURL := &models.ShortURL{
		ID:          shortid.Generate(),
		OriginalURL: OriginalURL,
		UserID:      userID,
	}
	if err = s.repo.ShortURLCreate(shortURL); err != nil {
		return nil, ErrInternal
	}

	// Возвращаем модель
	return shortURL, nil
}

// GetByID - возвращает ShortURL по его id
func (s ShortURLService) GetByID(id string) (*models.ShortURL, error) {
	shortURL, err := s.repo.ShortURLGetByID(id)
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}
	return shortURL, nil
}

// GetByUserID - возвращает все ShortURL пользователя
func (s ShortURLService) GetByUserID(id uint) ([]models.ShortURL, error) {
	shortURLs, err := s.repo.ShortURLGetByUserID(id)
	if err == repo.ErrNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, ErrInternal
	}
	return shortURLs, nil
}

func (s ShortURLService) Resolve(id string) string {
	return s.cfg.BaseURL.String() + id
}

// validateURL - проверяет URL на максимальную длину и http/https-протокол
func (s ShortURLService) validateURL(rawURL string) error {
	// Проверка на максимальную длину URL
	if len(rawURL) > s.cfg.URLMaxLen {
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