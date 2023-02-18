package app

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog/log"

	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/http/handlers"
	"github.com/ofstudio/go-shortener/internal/http/middleware"
	"github.com/ofstudio/go-shortener/internal/providers"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// HTTPServer - HTTP-сервер
type HTTPServer struct {
	cfg *config.Config
	u   *usecases.Container
	p   *providers.Container
}

// NewHTTPServer - конструктор сервера
func NewHTTPServer(cfg *config.Config, u *usecases.Container, p *providers.Container) *HTTPServer {
	return &HTTPServer{cfg: cfg, u: u, p: p}
}

// Start - запуск сервера
func (s *HTTPServer) Start(ctx context.Context) error {

	// Создаём маршрутизатор
	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(log.Logger))

	// Middleware для декомпрессии и компрессии ответов.
	// Параметр minSize рекомендуется равным middleware.MTUSize.
	// Значение 0 означает сжатие ответов любой длины и используется в целях демонстрации.
	r.Use(middleware.Decompressor)
	r.Use(middleware.NewCompressor(0, gzip.BestSpeed).
		AddType("application/json").
		AddType("text/plain").
		AddType("text/html").Handler)

	// Middleware аутентификационной куки.
	r.Use(s.p.Auth.Handler)

	// Публичные HTTP-запросы
	r.Mount("/", handlers.NewHTTPHandlers(s.u).Routes())

	// Публичный API
	apiHandlers := handlers.NewAPIHandlers(s.u)
	r.Mount("/api/", apiHandlers.PublicRoutes())

	// Внутренний API
	r.Group(func(r chi.Router) {
		// Middleware для проверки IP-адреса
		r.Use(s.p.IPCheck.Handler)
		r.Mount("/api/internal/", apiHandlers.InternalRoutes())
	})

	// Создаём HTTP-сервер
	server := &http.Server{
		Addr:    s.cfg.HTTPServerAddress,
		Handler: r,
	}

	// Горутина для остановки HTTP-сервера
	stop := make(chan error)
	go func() {
		<-ctx.Done()                                  // Ждём сигнала остановки
		stop <- server.Shutdown(context.Background()) // Останавливаем сервер
	}()

	// Создаем листенер для HTTP-сервера
	l, err := listener(s.cfg.HTTPServerAddress, s.cfg.EnableHTTPS, s.p.TLSConf)
	if err != nil {
		return fmt.Errorf("failed to create HTTP listener: %w", err)
	}
	//goland:noinspection ALL
	defer l.Close()

	// Запускаем HTTP-сервер
	log.Info().Msgf("Starting HTTP server on %s", l.Addr().String())
	err = server.Serve(l)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Ждём сигнала остановки HTTP-сервера, возвращаем результат остановки
	log.Warn().Msg("HTTP server stopped")
	return <-stop
}
