package server

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

	"github.com/ofstudio/go-shortener/internal/app/config"
)

// selfSignedTLS - возвращает конфигурацию TLS c самоподписанным сертификатом.
func selfSignedTLS(c *config.TLS) (*tls.Config, error) {
	cert, err := newSelfSignedCert(c)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

// newSelfSignedCert - возвращает самоподписанный сертификат.
func newSelfSignedCert(c *config.TLS) (tls.Certificate, error) {
	var curve elliptic.Curve
	switch c.Curve {
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
		Subject:               c.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(c.TTL),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	// добавляем хосты в сертификат
	for _, h := range c.Hosts {
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
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err != nil {
		return tls.Certificate{}, err
	}

	// возвращаем tls.Certificate
	return tls.X509KeyPair(certPEM.Bytes(), privateKeyPEM.Bytes())
}
