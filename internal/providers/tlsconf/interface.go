package tlsconf

import (
	"crypto/tls"
)

// Provider - интерфейс провайдера конфигурации TLS.
type Provider interface {
	// Config - создает net.Listener для указанного адреса.
	Config() (*tls.Config, error)
}
