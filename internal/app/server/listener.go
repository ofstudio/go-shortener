package server

import (
	"crypto/tls"
	"net"

	"github.com/ofstudio/go-shortener/internal/app/config"
)

// NewListener - возвращает net.Listener с TLS c самоподписанным сертификатом,
// либо без TLS.
func NewListener(cfg *config.Config) (net.Listener, error) {
	if cfg.EnableHTTPS {
		c, err := selfSignedTLS(&cfg.TLS)
		if err != nil {
			return nil, err
		}
		return tls.Listen("tcp", cfg.ServerAddress, c)
	}
	return net.Listen("tcp", cfg.ServerAddress)
}
