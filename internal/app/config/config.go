package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"net/url"
	"os"
	"time"
)

// Config - конфигурация приложения
type Config struct {
	// BaseURL - базовый адрес сокращённого URL.
	BaseURL url.URL `env:"BASE_URL"`

	// ServerAddress - адрес для запуска HTTP-сервера.
	ServerAddress string `env:"SERVER_ADDRESS"`

	// FileStoragePath - файл для хранения данных.
	// Если не задан, данные будут храниться в памяти.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`

	// AuthTTL - время жизни авторизационного токена
	AuthTTL time.Duration `env:"AUTH_TTL"`

	// AuthSecret - секретный ключ для подписи авторизационного токена
	AuthSecret string `env:"AUTH_SECRET,unset"`

	// DatabaseDSN - строка с адресом подключения к БД
	DatabaseDSN string `env:"DATABASE_DSN"`
}

var defaultConfig = Config{
	BaseURL:       mustParseRequestURI("http://localhost:8080/"),
	ServerAddress: "0.0.0.0:8080",
	AuthTTL:       time.Minute * 60 * 24 * 30,
	AuthSecret:    mustRandSecret(64),
	DatabaseDSN:   "",
}

// Default - конфигурационная функция, которая возвращает конфигурацию по умолчанию.
// Входной параметр не используется. Ошибки не возвращаются.
func Default(_ *Config) (*Config, error) {
	cfg := defaultConfig
	return &cfg, nil
}

// FromCLI - конфигурационная функция, которая считывает конфигурацию приложения из переменных окружения.
//
// Флаги командной строки:
//    -a <host:port> - адрес для запуска HTTP-сервера
//    -b <url>       - базовый адрес сокращённого URL
//    -f <path>      - файл для хранения данных
//    -t <duration>  - время жизни авторизационного токена
//	  -d <dsn>       - строка с адресом подключения к БД
//
// Если какие-либо значения не заданы в командной строке, то используются значения переданные в cfg.
func FromCLI(cfg *Config) (*Config, error) {
	return fromCLI(cfg, os.Args[1:]...)
}

// fromCLI - логика для NewFromEnvAndCLI.
// Вынесена отдельно в целях тестирования.
func fromCLI(cfg *Config, arguments ...string) (*Config, error) {
	// Парсим командную строку
	cli := flag.NewFlagSet("config", flag.ExitOnError)
	cli.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	cli.Func("b", "Base URL", urlParseFunc(&cfg.BaseURL))
	cli.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path (default: in-memory)")
	cli.DurationVar(&cfg.AuthTTL, "t", cfg.AuthTTL, "Auth token TTL")
	cli.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	if err := cli.Parse(arguments); err != nil {
		return nil, err
	}
	return validate(cfg)
}

// FromEnv - конфигурационная функция, которая читывает конфигурацию приложения из переменных окружения.
//
// Переменные окружения:
//    SERVER_ADDRESS    - адрес для запуска HTTP-сервера
//    BASE_URL          - базовый адрес сокращённого URL
//    FILE_STORAGE_PATH - файл для хранения данных
//    AUTH_TTL          - время жизни авторизационного токена
//	  AUTH_SECRET       - секретный ключ для подписи авторизационного токена
//
// Если какие-либо переменные окружения не заданы, то используются значения переданные в cfg.
func FromEnv(cfg *Config) (*Config, error) {
	// Получаем параметры из окружения
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return validate(cfg)
}
