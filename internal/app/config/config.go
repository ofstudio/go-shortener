package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"golang.org/x/sync/errgroup"
	"net"
	"net/url"
	"os"
	"strings"
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

// Default - конфигурационная функция, которая возвращает конфигурацию по умолчанию.
// Входной параметр не используется. Ошибки не возвращаются.
func Default(_ *Config) (*Config, error) {
	secret, err := randSecret(64)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		BaseURL:       url.URL{Scheme: "http", Host: "localhost:8080", Path: "/"},
		ServerAddress: "0.0.0.0:8080",
		AuthTTL:       time.Minute * 60 * 24 * 30,
		AuthSecret:    secret,
		DatabaseDSN:   "",
	}
	if err = cfg.validate(); err != nil {
		return nil, err
	}
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

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
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
	if err = cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate - проверяет конфигурацию на валидность
func (c *Config) validate() error {
	g := &errgroup.Group{}
	g.Go(c.validateAuthSecret)
	g.Go(c.validateBaseURL)
	g.Go(c.validateServerAddr)
	return g.Wait()
}

// validateBaseURL - проверяет базовый адрес сокращённого URL.
// Возвращает ошибку в случае:
//    - URL не содержит протокол http или https
//    - URL содержит параметры или фрагмент.
// Добавляет слеш в конце Path, если его нет.
func (c *Config) validateBaseURL() error {
	if c.BaseURL.RawQuery != "" || c.BaseURL.Fragment != "" {
		return fmt.Errorf("base URL must not contain query parameters or fragment")
	}
	if c.BaseURL.Scheme != "http" && c.BaseURL.Scheme != "https" {
		return fmt.Errorf("base URL must use http or https scheme")
	}
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		c.BaseURL.Path += "/"
	}
	return nil
}

// validateServerAddr - проверяет адрес для запуска HTTP-сервера.
func (c *Config) validateServerAddr() error {
	if c.ServerAddress == "" {
		return fmt.Errorf("empty server address")
	}
	_, err := net.ResolveTCPAddr("tcp", c.ServerAddress)
	if err != nil {
		return fmt.Errorf("invalid server address")
	}
	return nil
}

// validateAuthSecret - проверяет чтобы ключ авторизации был не пустым.
func (c *Config) validateAuthSecret() error {
	if len(c.AuthSecret) == 0 {
		return fmt.Errorf("auth secret not set")
	}
	return nil
}
