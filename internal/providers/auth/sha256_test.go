package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/internal/usecases"
	"github.com/ofstudio/go-shortener/test/pbsuite"
)

type sha256ProviderSuite struct {
	suite.Suite
	cfg *config.Config
	u   *usecases.User
}

func TestSHA256ProviderSuite(t *testing.T) {
	suite.Run(t, new(sha256ProviderSuite))
}

func (suite *sha256ProviderSuite) SetupTest() {
	suite.cfg = &config.Config{
		AuthSecret: "secret",
		AuthTTL:    time.Hour,
	}
	suite.u = usecases.NewUser(repo.NewMemoryRepo())
}

func (suite *sha256ProviderSuite) TestCreateToken() {
	provider := NewSHA256Provider(suite.cfg, suite.u)
	token, err := provider.CreateToken(1)
	suite.NoError(err)
	suite.NotEmpty(token)
}

func (suite *sha256ProviderSuite) TestVerifyToken() {
	suite.Run("valid token", func() {
		provider := NewSHA256Provider(suite.cfg, suite.u)
		token, err := provider.CreateToken(100)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(token)

		userID, err := provider.VerifyToken(token)
		suite.NoError(err)
		suite.Equal(uint(100), userID)
	})

	suite.Run("invalid secret", func() {
		suite.cfg.AuthSecret = "invalid"
		provider := NewSHA256Provider(suite.cfg, suite.u)
		token, err := provider.CreateToken(100)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(token)

		suite.cfg.AuthSecret = "secret"
		provider = NewSHA256Provider(suite.cfg, suite.u)
		_, err = provider.VerifyToken(token)
		suite.ErrorIs(err, ErrInvalidToken)
	})

	suite.Run("expired token", func() {
		suite.cfg.AuthTTL = -time.Minute
		provider := NewSHA256Provider(suite.cfg, suite.u)
		token, err := provider.CreateToken(100)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(token)

		_, err = provider.VerifyToken(token)
		suite.ErrorIs(err, ErrExpiredToken)
	})

	suite.Run("invalid token", func() {
		provider := NewSHA256Provider(suite.cfg, suite.u)
		_, err := provider.VerifyToken("0102030405")
		suite.ErrorIs(err, ErrInvalidToken)
	})
}

type sha256ProviderHTTPSuite struct {
	suite.Suite
	cfg      *config.Config
	u        *usecases.User
	testHTTP *httptest.Server
}

func TestSHA256ProviderHTTPSuite(t *testing.T) {
	suite.Run(t, new(sha256ProviderHTTPSuite))
}

func (suite *sha256ProviderHTTPSuite) SetupTest() {
	suite.cfg = &config.Config{
		AuthSecret: "secret",
		AuthTTL:    time.Hour,
	}
	suite.u = usecases.NewUser(repo.NewMemoryRepo())

	r := chi.NewRouter()
	r.Use(NewSHA256Provider(suite.cfg, suite.u).Handler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	suite.testHTTP = httptest.NewServer(r)
}

func (suite *sha256ProviderHTTPSuite) TearDownTest() {
	suite.testHTTP.Close()
}

func (suite *sha256ProviderHTTPSuite) TestHandler_CreateUser() {
	suite.Run("should create new user with token", func() {
		n, err := suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(0, n)

		resp, err := http.Get(suite.testHTTP.URL)
		suite.Require().NoError(err)
		defer suite.NoError(resp.Body.Close())
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
		suite.Require().Len(resp.Cookies(), 1)
		suite.Require().Equal("auth_token", resp.Cookies()[0].Name)
		suite.Require().NotEmpty(resp.Cookies()[0].Value)
		suite.Require().Equal(int(suite.cfg.AuthTTL/time.Second), resp.Cookies()[0].MaxAge)

		n, err = suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(1, n)
	})
}
func (suite *sha256ProviderHTTPSuite) TestHandler_AcceptToken() {
	suite.Run("should accept valid token", func() {
		n, err := suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(0, n)

		resp, err := http.Get(suite.testHTTP.URL)
		suite.Require().NoError(err)
		defer suite.NoError(resp.Body.Close())
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
		suite.Require().Len(resp.Cookies(), 1)
		suite.Require().Equal("auth_token", resp.Cookies()[0].Name)
		c := resp.Cookies()[0]

		n, err = suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(1, n)

		req, err := http.NewRequest(http.MethodGet, suite.testHTTP.URL, nil)
		suite.Require().NoError(err)
		req.AddCookie(c)
		resp, err = http.DefaultClient.Do(req)
		suite.Require().NoError(err)
		defer suite.NoError(resp.Body.Close())

		n, err = suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(1, n)
	})
}

type sha256ProviderGRPSSuite struct {
	pbsuite.Suite
	cfg *config.Config
	u   *usecases.User
}

func TestSHA256ProviderGRPSSuite(t *testing.T) {
	suite.Run(t, new(sha256ProviderGRPSSuite))
}

func (suite *sha256ProviderGRPSSuite) SetupTest() {
	suite.cfg = &config.Config{
		AuthSecret: "secret",
		AuthTTL:    time.Hour,
	}
	suite.u = usecases.NewUser(repo.NewMemoryRepo())
	suite.StartServer(grpc.NewServer(
		grpc.UnaryInterceptor(NewSHA256Provider(suite.cfg, suite.u).Interceptor),
	))
}

func (suite *sha256ProviderGRPSSuite) TestInterceptor_CreateUser() {

	suite.Run("should create new user with token", func() {
		n, err := suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(0, n)

		var header metadata.MD
		_, err = suite.HelloClient.Hello(context.Background(), &pbsuite.Empty{}, grpc.Header(&header))
		suite.Require().NoError(err)
		suite.Require().Len(header.Get("auth_token"), 1)
		suite.Require().NotEmpty(header.Get("auth_token")[0])

		n, err = suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(1, n)
	})
}

func (suite *sha256ProviderGRPSSuite) TestInterceptor_AcceptToken() {

	suite.Run("should accept valid token", func() {
		n, err := suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(0, n)

		var header metadata.MD
		_, err = suite.HelloClient.Hello(context.Background(), &pbsuite.Empty{}, grpc.Header(&header))

		suite.Require().NoError(err)
		suite.Require().Len(header.Get("auth_token"), 1)
		suite.Require().NotEmpty(header.Get("auth_token")[0])
		token := header.Get("auth_token")[0]

		n, err = suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(1, n)

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("auth_token", token))
		_, err = suite.HelloClient.Hello(ctx, &pbsuite.Empty{}, grpc.Header(&header))
		suite.Require().NoError(err)

		n, err = suite.u.Count(context.Background())
		suite.Require().NoError(err)
		suite.Require().Equal(1, n)
	})
}
