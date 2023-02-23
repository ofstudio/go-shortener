package config

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"time"
)

// Cert - конфигурация для самоподписанного сертификата
type Cert struct {
	// Hosts - список хостов для сертификата
	Hosts []string
	// Subject - информация о владельце сертификата
	Subject pkix.Name
	// Curve - кривая для генерации ключа
	Curve tls.CurveID
	// TTL - время жизни сертификата
	TTL time.Duration
}

// defaultCert - конфигурация Cert по умолчанию для самоподписанного сертификата
var defaultCert = Cert{
	Hosts: []string{"localhost", "127.0.0.1"},
	Subject: pkix.Name{
		Organization: []string{"Yandex.Praktikum"},
		Country:      []string{"RU"},
	},
	Curve: tls.CurveP256,
	TTL:   time.Hour * 24 * 365 * 10, // 10 лет
}

// validate - проверка конфигурации Cert
func (c *Cert) validate() error {
	if len(c.Hosts) == 0 {
		return fmt.Errorf("empty hosts list")
	}
	if c.TTL <= 0 {
		return fmt.Errorf("invalid TTL: %v", c.TTL)
	}
	return nil
}
