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

	// HTTPServerAddress - адрес для запуска HTTP-сервера.
	HTTPServerAddress string `env:"SERVER_ADDRESS"`

	// GRPCServerAddress - адрес для запуска GRPC-сервера.
	GRPCServerAddress string `env:"GRPC_SERVER_ADDRESS"`

	// FileStoragePath - файл для хранения данных
	FileStoragePath string `env:"FILE_STORAGE_PATH"`

	// DatabaseDSN - строка с адресом подключения к БД
	DatabaseDSN string `env:"DATABASE_DSN"`

	// AuthSecret - секретный ключ для подписи авторизационного токена
	AuthSecret string `env:"AUTH_SECRET,unset"`

	// TrustedSubnet - подсеть, из которой разрешено обращение к внутреннему API
	TrustedSubnet string `env:"TRUSTED_SUBNET"`

	// configFName - имя файла конфигурации
	configFName string

	Cert Cert

	// EnableHTTPS - использовать самоподписный Cert
	EnableHTTPS bool `env:"ENABLE_HTTPS"`

	// AuthTTL - время жизни авторизационного токена
	AuthTTL time.Duration `env:"AUTH_TTL"`
}

// validate - проверяет конфигурацию на валидность
func (c *Config) validate() error {
	g := &errgroup.Group{}
	g.Go(c.validateAuthSecret)
	g.Go(c.validateBaseURL)
	g.Go(c.validateServerAddr)
	g.Go(c.Cert.validate)
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
	if c.HTTPServerAddress == "" {
		return fmt.Errorf("empty HTTP server address")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.HTTPServerAddress); err != nil {
		return fmt.Errorf("invalid HTTP server address")
	}

	if c.GRPCServerAddress == "" {
		return fmt.Errorf("empty GRPC server address")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.GRPCServerAddress); err != nil {
		return fmt.Errorf("invalid GRPC server address")
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
