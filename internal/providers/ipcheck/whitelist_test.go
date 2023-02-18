package ipcheck

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ofstudio/go-shortener/test/pbsuite"
)

type whitelistSuite struct {
	suite.Suite
}

func TestWhitelistSuite(t *testing.T) {
	suite.Run(t, new(whitelistSuite))
}

func (suite *whitelistSuite) TestNewWhitelistIsAllowed() {
	suite.Run("empty whitelist", func() {
		wl := NewWhitelist()
		suite.False(wl.IsAllowed(net.IP{127, 0, 0, 1}))
	})

	suite.Run("whitelist with single IP", func() {
		wl := NewWhitelist("127.0.0.1")
		suite.True(wl.IsAllowed(net.IP{127, 0, 0, 1}))
		suite.False(wl.IsAllowed(net.IP{127, 0, 0, 2}))
	})

	suite.Run("whitelist with multiple IPs", func() {
		wl := NewWhitelist("127.0.0.1", "127.0.0.2")
		suite.True(wl.IsAllowed(net.IP{127, 0, 0, 1}))
		suite.True(wl.IsAllowed(net.IP{127, 0, 0, 2}))
		suite.False(wl.IsAllowed(net.IP{127, 0, 0, 3}))
	})

	suite.Run("whitelist with single CIDR", func() {
		wl := NewWhitelist("192.168.0.0/24")
		suite.True(wl.IsAllowed(net.IP{192, 168, 0, 1}))
		suite.True(wl.IsAllowed(net.IP{192, 168, 0, 255}))
		suite.False(wl.IsAllowed(net.IP{192, 168, 1, 1}))
	})

	suite.Run("whitelist with multiple CIDRs", func() {
		wl := NewWhitelist("192.168.0.0/24", "10.10.0.0/16")
		suite.True(wl.IsAllowed(net.IP{192, 168, 0, 1}))
		suite.True(wl.IsAllowed(net.IP{192, 168, 0, 255}))
		suite.False(wl.IsAllowed(net.IP{192, 168, 1, 1}))
		suite.True(wl.IsAllowed(net.IP{10, 10, 0, 1}))
		suite.True(wl.IsAllowed(net.IP{10, 10, 255, 255}))
		suite.False(wl.IsAllowed(net.IP{10, 11, 0, 1}))
	})

	suite.Run("whitelist with mixed IPs and CIDRs", func() {
		wl := NewWhitelist("127.0.0.1", "192.168.0.0/24")
		suite.True(wl.IsAllowed(net.IP{127, 0, 0, 1}))
		suite.False(wl.IsAllowed(net.IP{127, 0, 0, 2}))
		suite.True(wl.IsAllowed(net.IP{192, 168, 0, 1}))
		suite.True(wl.IsAllowed(net.IP{192, 168, 0, 255}))
		suite.False(wl.IsAllowed(net.IP{192, 168, 1, 1}))
	})

	suite.Run("whitelist with invalid addresses", func() {
		wl := NewWhitelist("localhost")
		suite.False(wl.IsAllowed(net.IP{127, 0, 0, 1}))
	})
}

type whitelistHTTPSuite struct {
	suite.Suite
	testHTTP *httptest.Server
}

func TestWhitelistHTTPSuite(t *testing.T) {
	suite.Run(t, new(whitelistHTTPSuite))
}

func (suite *whitelistHTTPSuite) SetupTest() {
	r := chi.NewRouter()
	r.With(NewWhitelist("127.0.0.1").Handler).
		Get("/x-real-ip", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	r.With(NewWhitelist("127.0.0.1").UseHeaders("X-Forwarded-For").Handler).
		Get("/x-forwarded-for", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	suite.testHTTP = httptest.NewServer(r)
}

func (suite *whitelistHTTPSuite) TearDownTest() {
	suite.testHTTP.Close()
}

func (suite *whitelistHTTPSuite) TestHandler() {

	suite.Run("should pass with allowed address in X-Real-IP", func() {
		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL+"/x-real-ip", nil)
		suite.NoError(err)
		req.Header.Set("X-Real-IP", "127.0.0.1")
		resp, err := http.DefaultClient.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusOK, resp.StatusCode)
	})

	suite.Run("should fail with denied address in X-Real-IP", func() {
		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL+"/x-real-ip", nil)
		suite.NoError(err)
		req.Header.Set("X-Real-IP", "127.0.0.2")
		resp, err := http.DefaultClient.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusForbidden, resp.StatusCode)
	})

	suite.Run("should fail with invalid address in X-Real-IP", func() {
		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL+"/x-real-ip", nil)
		suite.NoError(err)
		req.Header.Set("X-Real-IP", "localhost")
		resp, err := http.DefaultClient.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusForbidden, resp.StatusCode)
	})

	suite.Run("should fail without X-Real-IP", func() {
		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL+"/x-real-ip", nil)
		suite.NoError(err)
		resp, err := http.DefaultClient.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusForbidden, resp.StatusCode)
	})

	suite.Run("should pass with allowed address in X-Forwarded-For", func() {
		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL+"/x-forwarded-for", nil)
		suite.NoError(err)
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		resp, err := http.DefaultClient.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusOK, resp.StatusCode)
	})

	suite.Run("should fail with denied address in X-Forwarded-For", func() {
		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL+"/x-forwarded-for", nil)
		suite.NoError(err)
		req.Header.Set("X-Forwarded-For", "127.0.0.2")
		resp, err := http.DefaultClient.Do(req)
		suite.NoError(err)
		suite.Equal(http.StatusForbidden, resp.StatusCode)
	})
}

type whitelistGRPCSuite struct {
	pbsuite.Suite
	wl *Whitelist
}

func TestWhitelistGRPCSuite(t *testing.T) {
	suite.Run(t, new(whitelistGRPCSuite))
}

func (suite *whitelistGRPCSuite) SetupSuite() {
	suite.wl = NewWhitelist("127.0.0.1")
	suite.StartServer(grpc.NewServer(
		grpc.UnaryInterceptor(suite.wl.Interceptor),
	))
}

func (suite *whitelistGRPCSuite) TestInterceptor() {

	suite.Run("should pass with allowed address in X-Real-IP", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.Pairs("X-Real-IP", "127.0.0.1"),
		)
		_, err := suite.HelloClient.Hello(ctx, &pbsuite.Empty{})
		suite.NoError(err)
	})

	suite.Run("should fail with denied address in X-Real-IP", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.Pairs("X-Real-IP", "127.0.0.2"),
		)
		_, err := suite.HelloClient.Hello(ctx, &pbsuite.Empty{})
		suite.Error(err)
		suite.Equal(codes.PermissionDenied, status.Code(err))
	})

	suite.Run("should fail with invalid address in X-Real-IP", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.Pairs("X-Real-IP", "localhost"),
		)
		_, err := suite.HelloClient.Hello(ctx, &pbsuite.Empty{})
		suite.Error(err)
		suite.Equal(codes.PermissionDenied, status.Code(err))
	})

	suite.Run("should fail without X-Real-IP", func() {
		_, err := suite.HelloClient.Hello(context.Background(), &pbsuite.Empty{})
		suite.Error(err)
		suite.Equal(codes.PermissionDenied, status.Code(err))
	})

	suite.Run("should pass with allowed address in X-Forwarded-For", func() {
		suite.wl.UseHeaders("X-Forwarded-For")
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.Pairs("X-Forwarded-For", "127.0.0.1"),
		)
		_, err := suite.HelloClient.Hello(ctx, &pbsuite.Empty{})
		suite.NoError(err)
	})
}
