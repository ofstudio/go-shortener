package config

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"time"
)

// TLS - конфигурация TLS для самоподписанного сертификата
type TLS struct {
	// Hosts - список хостов для сертификата
	Hosts []string
	// Subject - информация о владельце сертификата
	Subject pkix.Name
	// Curve - кривая для генерации ключа
	Curve tls.CurveID
	// TTL - время жизни сертификата
	TTL time.Duration
}

// DefaultTLS - конфигурация TLS по умолчанию для самоподписанного сертификата
var tlsDefault = TLS{
	Hosts: []string{"localhost", "127.0.0.1"},
	Subject: pkix.Name{
		Organization: []string{"Yandex.Praktikum"},
		Country:      []string{"RU"},
	},
	Curve: tls.CurveP256,
	TTL:   time.Hour * 24 * 365 * 10, // 10 лет
}

// validate - проверка конфигурации TLS
func (t *TLS) validate() error {
	if len(t.Hosts) == 0 {
		return fmt.Errorf("empty hosts list")
	}
	if t.TTL <= 0 {
		return fmt.Errorf("invalid TTL: %v", t.TTL)
	}
	return nil
}
