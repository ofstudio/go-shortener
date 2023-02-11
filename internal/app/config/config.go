package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
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

	// DatabaseDSN - строка с адресом подключения к БД
	DatabaseDSN string `env:"DATABASE_DSN"`

	// AuthSecret - секретный ключ для подписи авторизационного токена
	AuthSecret string `env:"AUTH_SECRET,unset"`

	// TrustedSubnet - подсеть, из которой разрешено обращение к внутреннему API
	TrustedSubnet string `env:"TRUSTED_SUBNET"`

	// configFName - имя файла конфигурации
	configFName string

	TLS TLS

	// EnableHTTPS - использовать самоподписный TLS
	EnableHTTPS bool `env:"ENABLE_HTTPS"`

	// AuthTTL - время жизни авторизационного токена
	AuthTTL time.Duration `env:"AUTH_TTL"`
}

// Default - конфигурационная функция, которая возвращает конфигурацию по умолчанию.
// Входной параметр не используется.
func Default(_ *Config) (*Config, error) {
	secret, err := randSecret(64)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		BaseURL:       url.URL{Scheme: "http", Host: "localhost:8080", Path: "/"},
		ServerAddress: "0.0.0.0:8080",
		EnableHTTPS:   false,
		TLS:           tlsDefault,
		AuthTTL:       time.Minute * 60 * 24 * 30,
		AuthSecret:    secret,
		DatabaseDSN:   "",
	}
	if err = cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// validate - проверяет конфигурацию на валидность
func (c *Config) validate() error {
	g := &errgroup.Group{}
	g.Go(c.validateAuthSecret)
	g.Go(c.validateBaseURL)
	g.Go(c.validateServerAddr)
	g.Go(c.TLS.validate)
	return g.Wait()
}

// validateBaseURL - проверяет базовый адрес сокращённого URL.
// Возвращает ошибку в случае:
//   - URL не содержит протокол http или https
//   - URL содержит параметры или фрагмент.
//
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
