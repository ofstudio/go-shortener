package app

import (
	"context"
	"fmt"

	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/ofstudio/go-shortener/api/proto"
	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/grpc/interceptors"
	"github.com/ofstudio/go-shortener/internal/grpc/services"
	"github.com/ofstudio/go-shortener/internal/providers"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// GRPCServer - gRPC-сервер
type GRPCServer struct {
	cfg *config.Config
	u   *usecases.Container
	p   *providers.Container
}

// NewGRPCServer - конструктор GRPCServer
func NewGRPCServer(cfg *config.Config, u *usecases.Container, p *providers.Container) *GRPCServer {
	return &GRPCServer{cfg: cfg, u: u, p: p}
}

// Start - запуск сервера
func (s *GRPCServer) Start(ctx context.Context) error {
	// Создаём gRPC-сервер
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(log.Logger)),
			s.p.Auth.Interceptor,
			interceptors.With("proto.Internal", s.p.IPCheck.Interceptor),
		),
	)
	// Регистрируем сервисы
	proto.RegisterShortURLServer(server, services.NewShortURLService(s.u))
	proto.RegisterInternalServer(server, services.NewInternalService(s.u))

	// Горутина для остановки gRPC-сервера
	stop := make(chan struct{})
	go func() {
		<-ctx.Done()          // Ждём сигнала остановки
		server.GracefulStop() // Останавливаем сервер
		close(stop)
	}()

	// Создаем листенер для gRPC-сервера
	l, err := listener(s.cfg.GRPCServerAddress, s.cfg.EnableHTTPS, s.p.TLSConf)
	if err != nil {
		return fmt.Errorf("failed to create gRPS listener: %w", err)
	}
	//goland:noinspection ALL
	defer l.Close()

	log.Info().Msgf("Starting gRPC server on %s", l.Addr().String())
	if err = server.Serve(l); err != nil {
		return err
	}

	log.Warn().Msg("gRPC server stopped")
	return nil
}
