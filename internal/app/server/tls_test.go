package server

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/stretchr/testify/suite"
)

type tlsSuite struct {
	suite.Suite
	server *httptest.Server
}

func TestTLSSuite(t *testing.T) {
	suite.Run(t, new(tlsSuite))
}

func (suite *tlsSuite) SetupTest() {
	suite.server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func (suite *tlsSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *tlsSuite) TestSelfSignedTLS() {
	tlsCfg, err := selfSignedTLS(&config.TLS{
		Hosts: []string{"localhost", "127.0.0.1"},
		Subject: pkix.Name{
			Organization: []string{"test"},
			Country:      []string{"RU"},
		},
		Curve: tls.CurveP256,
		TTL:   time.Hour,
	})
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
}
