package app

import (
	"crypto/tls"
	"net"

	"github.com/ofstudio/go-shortener/internal/providers/tlsconf"
)

// listener - создает TCP net.Listener для указанного адреса.
// Если установлен флаг useTLS, то создается net.Listener c TLS c конфигурацией из provider.
func listener(addr string, useTLS bool, provider tlsconf.Provider) (net.Listener, error) {
	if !useTLS {
		return net.Listen("tcp", addr)
	}

	tlsCfg, err := provider.Config()
	if err != nil {
		return nil, err
	}
	return tls.Listen("tcp", addr, tlsCfg)
}
