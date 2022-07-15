package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// validate - проверяет конфигурацию на валидность
func validate(cfg *Config) (*Config, error) {
	var err error
	// Валидируем базовый AuthSecret
	if err = validateAuthSecret(cfg.AuthSecret); err != nil {
		return nil, err
	}
	// Валидируем базовый URL
	if err = validateBaseURL(&cfg.BaseURL); err != nil {
		return nil, err
	}
	// Валидируем адрес для запуска HTTP-сервера
	if err = validateServerAddr(cfg.ServerAddress); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validateBaseURL - проверяет базовый адрес сокращённого URL.
// Возвращает ошибку в случае:
//    - URL не содержит протокол http или https
//    - URL содержит параметры или фрагмент.
// Добавляет слеш в конце Path, если его нет.
func validateBaseURL(baseURL *url.URL) error {
	if baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return fmt.Errorf("base URL must not contain query parameters or fragment")
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return fmt.Errorf("base URL must use http or https scheme")
	}
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return nil
}

// validateServerAddr - проверяет адрес для запуска HTTP-сервера.
func validateServerAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("empty server address")
	}
	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("invalid server address")
	}
	return nil
}

func validateAuthSecret(secret string) error {
	if len(secret) == 0 {
		return fmt.Errorf("auth secret not set")
	}
	return nil
}
