package services

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/storage"
	"github.com/ofstudio/go-shortener/pkg/shortid"
	"net/url"
)

type ShortenerService struct {
	cfg *config.Config
	db  storage.Interface
}

func NewShortenerService(cfg *config.Config, db storage.Interface) *ShortenerService {
	return &ShortenerService{cfg, db}
}

// CreateShortURL - создает и возвращает короткий URL
func (srv ShortenerService) CreateShortURL(longURL string) (string, error) {
	// Проверяем URL на валидность
	if err := srv.validateURL(longURL); err != nil {
		return "", err
	}

	// Генерируем id для URL и сохраняем URL в сторадж
	id := shortid.Generate()
	if srv.db.Set(id, longURL) != nil {
		return "", ErrInternal
	}

	// Возвращаем короткий URL
	return srv.cfg.PublicURL + id, nil
}

// GetLongURL - возвращает исходный URL по его id
func (srv ShortenerService) GetLongURL(id string) (string, error) {
	val, err := srv.db.Get(id)
	if err != nil {
		if storage.IsNotFound(err) {
			return "", ErrShortURLNotFound
		} else {
			return "", ErrInternal
		}
	}
	return val, nil
}

// validateURL - проверяет URL на максимальную длину и http/https-протокол
func (srv ShortenerService) validateURL(rawURL string) error {
	// Проверка на максимальную длину URL
	if len(rawURL) > srv.cfg.URLMaxLen {
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
