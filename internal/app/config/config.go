package config

import (
	"errors"
	"github.com/caarlos0/env/v6"
	"net/url"
)

// Config - конфигурация приложения
type Config struct {
	// Максимальная длина URL в байтах.
	// Формально, размер URL ничем не ограничен.
	// Разные версии разных браузеров имеют свои конкретные ограничения: от 2048 байт до мегабайт.
	// В случае нашего сервиса необходимо некое разумное ограничение
	URLMaxLen int

	// Публичный URL, по которому доступно приложение
	// Пример: https://example.com/ - обязателен слеш на конце
	BaseURL string `env:"BASE_URL"`

	// Адрес запуска HTTP-сервера
	ServerAddress string `env:"SERVER_ADDRESS"`
}

// Конфигурация по умолчанию
var defaultConfig = Config{
	URLMaxLen:     2048,
	BaseURL:       "http://localhost:8080/",
	ServerAddress: ":8080",
}

func NewFromEnv() (*Config, error) {
	// Параметры по умолчанию
	cfg := defaultConfig
	// Получаем параметры из окружения
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	// Нормализуем BaseURL
	if cfg.BaseURL, err = normalizeBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// normalizeBaseURL - нормализует публичный URL.
// Возвращает ошибку если URL содежит параметры параметры, а также если URL пустой или невалидный.
// Добавляет слеш в конце, если его нет.
func normalizeBaseURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", errors.New("empty base URL")
	}
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", errors.New("invalid base URL")
	}
	if u.RawQuery != "" {
		return "", errors.New("base URL must not contain query parameters")
	}
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	return baseURL, nil
}
