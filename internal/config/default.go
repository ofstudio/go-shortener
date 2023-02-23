package config

import (
	"net/url"
	"time"
)

// Default - конфигурационная функция, которая возвращает конфигурацию по умолчанию.
// Входной параметр не используется.
func Default(_ *Config) (*Config, error) {
	secret, err := randSecret(64)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		BaseURL:           url.URL{Scheme: "http", Host: "localhost:8080", Path: "/"},
		HTTPServerAddress: ":8080",
		GRPCServerAddress: ":9090",
		EnableHTTPS:       false,
		Cert:              defaultCert,
		AuthTTL:           time.Minute * 60 * 24 * 30,
		AuthSecret:        secret,
		DatabaseDSN:       "",
	}
	if err = cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
