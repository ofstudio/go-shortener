package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
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
	// Обязателен слеш на конце.
	BaseURL string `env:"BASE_URL"`

	// Адрес для запуска HTTP-сервера.
	ServerAddress string `env:"SERVER_ADDRESS"`

	// Файл для хранения данных.
	// Если не задан, данные будут храниться в памяти.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// DefaultConfig - конфигурация по умолчанию
var DefaultConfig = Config{
	URLMaxLen:     2048,
	BaseURL:       "http://localhost:8080/",
	ServerAddress: "0.0.0.0:8080",
}

// NewFromEnv - инициализирует конфигурацию приложения из переменных окружения.
// Переменные окружения:
//    SERVER_ADDRESS    - адрес для запуска HTTP-сервера
//    BASE_URL          - базовый адрес сокращённого URL
//    FILE_STORAGE_PATH - файл для хранения данных
//
// Если какие-либо переменные окружения не заданы,
// используются значения по умолчанию из DefaultConfig.
func NewFromEnv() (*Config, error) {
	return validate(fromEnv())
}

// NewFromEnvAndCLI - инициализирует конфигурацию приложения из переменных окружения и командной строки.
// Значения из командной строки перекрывают значения из окружения.
//
// Переменные окружения:
//    SERVER_ADDRESS    - адрес для запуска HTTP-сервера
//    BASE_URL          - базовый адрес сокращённого URL
//    FILE_STORAGE_PATH - файл для хранения данных
//
// Флаги коммандной строки:
//    -a <host:port> - адрес для запуска HTTP-сервера
//    -b <url>       - базовый адрес сокращённого URL
//    -f <path>      - файл для хранения данных
//
// Если какие-либо переменные окружения не заданы,
// используются значения по умолчанию из DefaultConfig.
func NewFromEnvAndCLI() (*Config, error) {
	return validate(fromEnvAndCLI(os.Args[1:]))
}

// fromEnv - логика для NewFromEnv.
func fromEnv() (*Config, error) {
	// Параметры по умолчанию
	cfg := DefaultConfig
	// Получаем параметры из окружения
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// fromEnvAndCLI - логика для NewFromEnvAndCLI.
func fromEnvAndCLI(arguments []string) (*Config, error) {
	cfg, err := fromEnv()
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
	return cfg, nil
}
