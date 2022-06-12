package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net/url"
	"os"
)

// Config - конфигурация приложения
type Config struct {
	// Максимальная длина URL в байтах.
	// Формально, размер URL ничем не ограничен.
	// Разные версии разных браузеров имеют свои конкретные ограничения: от 2048 байт до мегабайт.
	// В случае нашего сервиса необходимо некое разумное ограничение.
	URLMaxLen int

	// Базовый адрес сокращённого URL.
	// Можно задать переменной окружения `BASE_URL`.
	// Пример: https://example.com/ - обязателен слеш на конце.
	BaseURL string `env:"BASE_URL"`

	// Адрес для запуска HTTP-сервера.
	// Можно задать переменной окружения `SERVER_ADDRESS`.
	ServerAddress string `env:"SERVER_ADDRESS"`

	// Файл для хранения данных.
	// Можно задать переменной окружения `FILE_STORAGE_PATH`.
	// Если не задан, данные будут храниться в памяти.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// DefaultConfig - конфигурация по умолчанию
var DefaultConfig = Config{
	URLMaxLen:     2048,
	BaseURL:       "http://localhost:8080/",
	ServerAddress: ":8080",
}

// NewFromEnv - инициализирует конфигурацию приложения из переменных окружения.
//    BASE_URL - базовый адрес сокращённого URL.
//    SERVER_ADDRESS - адрес для запуска HTTP-сервера.
//    FILE_STORAGE_PATH - файл для хранения данных.
// Если какие-либо переменные окружения не заданы, используются значения по умолчанию DefaultConfig.
func NewFromEnv() (*Config, error) {
	// Параметры по умолчанию
	cfg := DefaultConfig
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

// NewFromEnvAndCLI - инициализирует конфигурацию приложения из переменных окружения и командной строки.
// Значения из командной строки перекрывают значения из окружения.
// Флаги:
//    -a <host:port> адрес для запуска HTTP-сервера
//    -b <url>       базовый адрес сокращённого URL
//    -f <path>      файл для хранения данных
// Если какие-либо переменные окружения или флаги не заданы, используются значения по умолчанию DefaultConfig.
func NewFromEnvAndCLI() (*Config, error) {
	return newFromEnvAndCLI(os.Args[1:])
}

// newFromEnvAndCLI - бизнес-логика для NewFromEnvAndCLI.
// Вынесена в отдельную функцию для целей тестирования.
func newFromEnvAndCLI(arguments []string) (*Config, error) {
	cfg, err := NewFromEnv()
	if err != nil {
		return nil, err
	}

	// Парсим командную строку
	cli := flag.NewFlagSet("config", flag.ExitOnError)
	cli.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	cli.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	cli.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path (default: in-memory)")
	if err = cli.Parse(arguments); err != nil {
		return nil, err
	}

	// Нормализуем BaseURL
	if cfg.BaseURL, err = normalizeBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	return cfg, nil
}

// normalizeBaseURL - нормализует публичный URL.
// Возвращает ошибку если URL содежит параметры параметры, а также если URL пустой или невалидный.
// Добавляет слеш в конце, если его нет.
func normalizeBaseURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("empty base URL")
	}
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL")
	}
	if u.RawQuery != "" {
		return "", fmt.Errorf("base URL must not contain query parameters")
	}
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	return baseURL, nil
}
