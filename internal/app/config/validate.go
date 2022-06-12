package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// validate - проверяет конфигурацию на валидность
func validate(cfg *Config, err error) (*Config, error) {
	// Если from...-функция завершилась с ошибкой, возвращаем ошибку
	if err != nil {
		return nil, err
	}
	// Валидируем базовый URL
	if err = validateBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	// Добавляем слеш в конец базового URL
	cfg.BaseURL = normalizeBaseURL(cfg.BaseURL)
	// Валидный адрес для запуска HTTP-сервера
	if err = validateServerAddr(cfg.ServerAddress); err != nil {
		return nil, err
	}
	return cfg, nil
}

// normalizeBaseURL - нормализует базовый адрес сокращённого URL.
// Добавляет слеш в конце, если его нет.
func normalizeBaseURL(baseURL string) string {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL
}

// validateBaseURL - проверяет базовый адрес сокращённого URL.
// Возвращает ошибку если URL содежит параметры параметры, а также если URL пустой или невалидный.
func validateBaseURL(baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("empty base URL")
	}
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL")
	}
	if u.RawQuery != "" {
		return fmt.Errorf("base URL must not contain query parameters")
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
