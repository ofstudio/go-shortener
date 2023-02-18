package tlsconf

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/go-shortener/internal/config"
)

type selfSignedProviderSuite struct {
	suite.Suite
	server  *httptest.Server
	certCfg config.Cert
}

func TestSelfSignedProviderSuite(t *testing.T) {
	suite.Run(t, new(selfSignedProviderSuite))
}

func (suite *selfSignedProviderSuite) SetupTest() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	suite.server = httptest.NewUnstartedServer(h)

	suite.certCfg = config.Cert{
		Hosts: []string{"localhost", "127.0.0.1"},
		Subject: pkix.Name{
			Organization: []string{"test"},
			Country:      []string{"RU"},
		},
		Curve: tls.CurveP256,
		TTL:   time.Hour,
	}
}

func (suite *selfSignedProviderSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *selfSignedProviderSuite) TestListen() {

	suite.Run("should establish TLS connection", func() {
		provider := NewSelfSignedProvider(suite.certCfg)
		suite.Require().NotNil(provider)
		tlsCfg, err := provider.Config()
		suite.Require().NoError(err)
		suite.server.TLS = tlsCfg
		suite.server.StartTLS()

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// Отключаем дефолтную проверку сертификата.
					// В данном случае мы проверяем сертификат самостоятельно в VerifyConnection.
					InsecureSkipVerify: true,
					VerifyConnection: func(cs tls.ConnectionState) error {
						suite.Len(cs.PeerCertificates, 1)
						suite.Equal("test", cs.PeerCertificates[0].Issuer.Organization[0])
						suite.Equal("RU", cs.PeerCertificates[0].Issuer.Country[0])
						suite.Equal("test", cs.PeerCertificates[0].Subject.Organization[0])
						suite.Equal("RU", cs.PeerCertificates[0].Subject.Country[0])
						suite.Equal("localhost", cs.PeerCertificates[0].DNSNames[0])
						suite.Equal(net.IP{127, 0, 0, 1}, cs.PeerCertificates[0].IPAddresses[0])
						return nil
					},
				},
			},
		}

		res, err := client.Get(suite.server.URL)
		if err != nil {
			suite.Require().NoError(err)
		}
		suite.NoError(res.Body.Close())
	})

	suite.Run("should return same certificate if called twice", func() {
		provider := NewSelfSignedProvider(suite.certCfg)
		suite.Require().NotNil(provider)

		tlsCfg1, err := provider.Config()
		suite.Require().NoError(err)

		tlsCfg2, err := provider.Config()
		suite.Require().NoError(err)

		suite.Equal(tlsCfg1.Certificates[0].Certificate[0], tlsCfg2.Certificates[0].Certificate[0])
	})
}
