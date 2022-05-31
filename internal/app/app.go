package app

import (
	"github.com/ofstudio/go-shortener/internal/shortid"
	"github.com/ofstudio/go-shortener/internal/storage"
	"net/url"
)

type App struct {
	cfg *Config
	db  storage.Interface
}

func NewApp(cfg *Config, db storage.Interface) *App {
	return &App{cfg, db}
}

// CreateShortURL - создает и возвращает короткий URL
func (a App) CreateShortURL(longURL string) (string, error) {
	// Проверяем URL на валидность
	if err := a.validateURL(longURL); err != nil {
		return "", err
	}

	// Генерируем id для URL и сохраняем URL в сторадж
	id := shortid.New()
	if a.db.Set(id, longURL) != nil {
		return "", ErrInternal
	}

	// Возвращаем короткий URL
	return a.cfg.publicURL + id, nil
}

// GetLongURL - возвращает исходный URL по его id
func (a App) GetLongURL(id string) (string, error) {
	val, err := a.db.Get(id)
	if err != nil {
		if storage.IsNotFound(err) {
			return "", ErrURLNotFound
		} else {
			return "", ErrInternal
		}
	}
	return val, nil
}

// validateURL - проверяет URL на максимальную длину и http/https-протокол
func (a App) validateURL(rawURL string) error {
	// Проверка на максимальную длину URL
	if len(rawURL) > a.cfg.urlMaxLen {
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
