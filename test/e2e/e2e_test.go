package e2e

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ofstudio/go-shortener/api/proto"
	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/config"
)

type e2eSuite struct {
	suite.Suite
	httpURL  string
	grpcConn *grpc.ClientConn
	stopChan chan error         // Канал для ожидания завершения сервера
	stopFunc context.CancelFunc // Функция для остановки сервера
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(e2eSuite))
}

func (suite *e2eSuite) SetupTest() {
	var err error

	// Контекст для остановки сервера
	suite.stopChan = make(chan error)
	ctx := context.Background()
	ctx, suite.stopFunc = context.WithCancel(ctx)

	// Конфигурация сервера
	cfg, _ := config.Default(nil)
	cfg.HTTPServerAddress = ":18080"
	cfg.BaseURL = url.URL{Scheme: "http", Host: "localhost:18080", Path: "/"}
	cfg.GRPCServerAddress = ":19090"
	cfg.TrustedSubnet = "192.168.0.0/24"

	// Точки подключения клиентов
	suite.httpURL = "http://localhost" + cfg.HTTPServerAddress
	suite.grpcConn, err = grpc.DialContext(
		ctx,
		cfg.GRPCServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	suite.Require().NoError(err)

	// Запуск сервера
	go func() {
		suite.stopChan <- app.NewApp(cfg).Start(ctx)
	}()
	time.Sleep(4 * time.Second)
}

func (suite *e2eSuite) TearDownTest() {
	time.Sleep(1 * time.Second)
	suite.stopFunc()
	suite.Require().NoError(<-suite.stopChan)
	suite.Require().NoError(suite.grpcConn.Close())
}

func (suite *e2eSuite) Test_HTTP_PublicAPI() {
	client := clientJar()

	suite.Run("POST /api/shorten", func() {
		body := strings.NewReader(`{"url":"https://google.com"}`)
		req, err := http.NewRequest("POST", suite.httpURL+"/api/shorten", body)
		suite.Require().NoError(err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusCreated, resp.StatusCode)
	})

	suite.Run("POST /api/shorten/batch", func() {
		body := strings.NewReader(`
			[
				{"correlation_id":"1","original_url":"https://apple.com"},
				{"correlation_id":"2","original_url":"https://yandex.ru"}
			]`)
		req, err := http.NewRequest("POST", suite.httpURL+"/api/shorten/batch", body)
		suite.Require().NoError(err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusCreated, resp.StatusCode)
	})

	suite.Run("DELETE /api/user/urls", func() {
		body := strings.NewReader(`["https://something.com"]`)
		req, err := http.NewRequest("DELETE", suite.httpURL+"/api/user/urls", body)
		suite.Require().NoError(err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusAccepted, resp.StatusCode)
		time.Sleep(100 * time.Millisecond)
	})

	suite.Run("GET /api/user/urls", func() {
		req, err := http.NewRequest("GET", suite.httpURL+"/api/user/urls", nil)
		suite.Require().NoError(err)
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
	})
}

func (suite *e2eSuite) Test_HTTP_InternalAPI() {
	client := clientJar()
	suite.Run("GET /api/internal/stats from allowed IP", func() {
		req, err := http.NewRequest("GET", suite.httpURL+"/api/internal/stats", nil)
		suite.Require().NoError(err)
		req.Header.Set("X-Real-IP", "192.168.0.1")
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
	})

	suite.Run("GET /api/internal/stats from denied IP", func() {
		req, err := http.NewRequest("GET", suite.httpURL+"/api/internal/stats", nil)
		suite.Require().NoError(err)
		req.Header.Set("X-Real-IP", "10.10.0.1")
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusForbidden, resp.StatusCode)
	})
}

func (suite *e2eSuite) Test_HTTP_Endpoints() {
	var shortURL string
	client := clientJar()

	suite.Run("POST /", func() {
		body := strings.NewReader(`https://amazon.com`)
		req, err := http.NewRequest("POST", suite.httpURL+"/", body)
		suite.Require().NoError(err)
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusCreated, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		suite.Require().NoError(err)
		defer suite.NoError(resp.Body.Close())
		shortURL = string(b)
	})

	suite.Run("GET shortURL", func() {
		req, err := http.NewRequest("GET", shortURL, nil)
		suite.Require().NoError(err)
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusTemporaryRedirect, resp.StatusCode)
		suite.Require().Equal("https://amazon.com", resp.Header.Get("Location"))
	})

	suite.Run("GET /ping", func() {
		req, err := http.NewRequest("GET", suite.httpURL+"/ping", nil)
		suite.Require().NoError(err)
		resp, err := client.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
	})

}

func (suite *e2eSuite) Test_GRPC_PublicAPI() {
	client := proto.NewShortURLClient(suite.grpcConn)
	shortURLs := make([]string, 3)
	var token string

	suite.Run("Shorten", func() {
		var header metadata.MD
		resp, err := client.Create(
			context.Background(),
			&proto.ShortURLCreateRequest{Url: "https://google.com"},
			grpc.Header(&header),
		)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(resp.Result)
		shortURLs[0] = resp.Result

		suite.Require().Len(header.Get("auth_token"), 1)
		suite.Require().NotEmpty(header.Get("auth_token")[0])
		token = header.Get("auth_token")[0]
	})

	suite.Run("CreateBatch", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("auth_token", token))
		resp, err := client.CreateBatch(ctx, &proto.ShortURLCreateBatchRequest{
			Items: []*proto.ShortURLCreateBatchRequest_Item{
				{CorrelationId: "1", OriginalUrl: "https://apple.com"},
				{CorrelationId: "2", OriginalUrl: "https://yandex.ru"},
			},
		})
		suite.Require().NoError(err)
		suite.Require().Len(resp.Items, 2)
		shortURLs[1] = resp.Items[0].ShortUrl
		shortURLs[2] = resp.Items[1].ShortUrl
	})

	suite.Run("DeleteBatch", func() {
		id1 := shortURLs[0][strings.LastIndex(shortURLs[0], "/")+1:]
		id2 := shortURLs[1][strings.LastIndex(shortURLs[1], "/")+1:]

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("auth_token", token))
		_, err := client.DeleteBatch(ctx, &proto.ShortURLDeleteBatchRequest{Items: []string{id1, id2}})
		suite.Require().NoError(err)
		time.Sleep(100 * time.Millisecond)
	})

	suite.Run("GetByUserID", func() {
		time.Sleep(100 * time.Millisecond)
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("auth_token", token))
		resp, err := client.GetByUserID(ctx, &proto.ShortURLGetByUserIDRequest{})
		suite.Require().NoError(err)
		suite.Require().Len(resp.Items, 1)
		suite.Require().Equal("https://yandex.ru", resp.Items[0].OriginalUrl)
	})
}

func (suite *e2eSuite) Test_GRPC_InternalAPI() {
	client := proto.NewInternalClient(suite.grpcConn)

	suite.Run("Stats with allowed IP", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-real-ip", "192.168.0.1"))
		resp, err := client.Stats(ctx, &proto.StatsRequest{})
		suite.Require().NoError(err)
		suite.Require().NotEmpty(resp)
	})

	suite.Run("Stats with denied IP", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-real-ip", "10.10.0.1"))
		_, err := client.Stats(ctx, &proto.StatsRequest{})
		suite.Require().Error(err)
		suite.Require().Equal(codes.PermissionDenied, status.Code(err))
	})
}
