// Package providers - провайдеры приложения
package providers

import (
	"github.com/ofstudio/go-shortener/internal/providers/auth"
	"github.com/ofstudio/go-shortener/internal/providers/ipcheck"
	"github.com/ofstudio/go-shortener/internal/providers/tlsconf"
)

// Container - контейнер провайдеров
type Container struct {
	// Auth - провайдер авторизации
	Auth auth.Provider
	// TLSConf - провайдер конфигурации TLS
	TLSConf tlsconf.Provider
	// IPCheck - провайдер проверки IP
	IPCheck ipcheck.Provider
}
