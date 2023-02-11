package server

import (
	"compress/gzip"
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// Server - HTTP-сервер
type Server struct {
	cfg    *config.Config
	server *http.Server
}

// NewServer - конструктор сервера
func NewServer(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

// Start - запуск сервера
func (s *Server) Start(ctx context.Context) error {
	// Создаём репозиторий и сервисы.
	repository, err := repo.Fabric(s.cfg)
	if err != nil {
		log.Fatal(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer repository.Close()

	// Создаем юзкейсы
	srv := usecases.NewContainer(s.cfg, repository)

	// Создаём маршрутизатор
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)

	// Middleware для декомпрессии и компрессии ответов.
	// Параметр minSize рекомендуется равным middleware.MTUSize.
	// Значение 0 означает сжатие ответов любой длины и используется в целях демонстрации.
	r.Use(middleware.Decompressor)
	r.Use(middleware.NewCompressor(0, gzip.BestSpeed).
		AddType("application/json").
		AddType("text/plain").
		AddType("text/html").Handler)

	// Middleware аутентификационной куки.
	r.Use(middleware.NewAuthCookie(srv).
		WithSecret([]byte(s.cfg.AuthSecret)).
		WithDomain(s.cfg.BaseURL.Host).
		WithTTL(s.cfg.AuthTTL).
		WithSecure(s.cfg.BaseURL.Scheme == "https").Handler)

	// Публичные HTTP-запросы
	r.Mount("/", handlers.NewHTTPHandlers(srv).Routes())

	// Публичный API
	apiHandlers := handlers.NewAPIHandlers(srv)
	r.Mount("/api/", apiHandlers.PublicRoutes())

	// Внутренний API
	r.Group(func(r chi.Router) {
		r.Use(middleware.NewWhitelist(s.cfg.TrustedSubnet).Handler)
		r.Mount("/api/internal/", apiHandlers.InternalRoutes())
	})

	// Создаём HTTP-сервер
	s.server = &http.Server{
		Addr:    s.cfg.ServerAddress,
		Handler: r,
	}

	// Горутина для остановки HTTP-сервера
	serverStopped := make(chan error)
	go func() {
		<-ctx.Done()
		serverStopped <- s.server.Shutdown(context.Background())
	}()

	// Запускаем сервер
	l, err := NewListener(s.cfg)
	if err != nil {
		return err
	}
	err = s.server.Serve(l)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Ждём сигнала остановки HTTP-сервера
	err = <-serverStopped

	return err
}
