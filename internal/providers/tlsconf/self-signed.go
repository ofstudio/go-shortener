package tlsconf

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/ofstudio/go-shortener/internal/config"
)

// SelfSignedProvider - реализация Provider с самоподписанным сертификатом.
type SelfSignedProvider struct {
	tlsCfg  *tls.Config
	certCfg config.Cert
}

// NewSelfSignedProvider - конструктор SelfSignedProvider.
func NewSelfSignedProvider(certCfg config.Cert) *SelfSignedProvider {
	return &SelfSignedProvider{
		certCfg: certCfg,
	}
}

// Config - возвращает конфигурацию TLS.
func (p *SelfSignedProvider) Config() (*tls.Config, error) {
	// Если конфигурация еще не создана, то создаем ее
	if p.tlsCfg == nil {
		cert, err := p.newSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("tlsconf: failed to generate self-signed certificate: %w", err)
		}
		p.tlsCfg = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	return p.tlsCfg, nil
}

// newSelfSignedCert - возвращает самоподписанный сертификат.
func (p *SelfSignedProvider) newSelfSignedCert() (tls.Certificate, error) {
	var curve elliptic.Curve
	switch p.certCfg.Curve {
	case tls.CurveP384:
		curve = elliptic.P384()
	case tls.CurveP521:
		curve = elliptic.P521()
	default:
		curve = elliptic.P256()
	}

	// Генерируем случайный серийный номер сертификата
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               p.certCfg.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(p.certCfg.TTL),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	// добавляем хосты в сертификат
	for _, h := range p.certCfg.Hosts {
		if ip := net.ParseIP(h); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, h)
		}
	}

	// Генерируем ключ
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	// кодируем сертификат и ключ в формат PEM,
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return tls.Certificate{}, err
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	var privateKeyPEM bytes.Buffer
	if err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}); err != nil {
		return tls.Certificate{}, err
	}

	// возвращаем tls.Cert
	return tls.X509KeyPair(certPEM.Bytes(), privateKeyPEM.Bytes())
}
