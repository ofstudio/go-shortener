package main

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ofstudio/go-shortener/internal/app/listener"

	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"github.com/ofstudio/go-shortener/internal/repo"
)

func main() {

	// Выводим информацию о сборке
	fmt.Print(buildInfo())

	// Считываем конфигурацию
	cfg, err := config.Compose(
		config.Default,                   // Значения по умолчанию
		config.FromJSONFile(os.Args[1:]), // Значения из JSON-файла
		config.FromEnv,                   // Значения из переменных окружения
		config.FromCLI(os.Args[1:]),      // Значения из флагов командной строки
	)
	if err != nil {
		log.Fatal(err)
	}
	// Создаём репозиторий и сервисы.
	repository, err := repo.Fabric(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer repository.Close()
	srv := services.NewContainer(cfg, repository)

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
		WithSecret([]byte(cfg.AuthSecret)).
		WithDomain(cfg.BaseURL.Host).
		WithTTL(cfg.AuthTTL).
		WithSecure(cfg.BaseURL.Scheme == "https").Handler)

	// Добавляем рауты для обработки запросов.
	r.Mount("/", handlers.NewHTTPHandlers(srv).Routes())
	r.Mount("/api/", handlers.NewAPIHandlers(srv).Routes())

	// Создаём сервер.
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	// Горутина для graceful-остановки сервера.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("Stopping http server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// Запускаем сервер.
	if cfg.EnableHTTPS {
		log.Printf("Starting https server at %s", cfg.ServerAddress)
	} else {
		log.Printf("Starting http server at %s", cfg.ServerAddress)
	}
	l, err := listener.NewListener(cfg)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
	err = server.Serve(l)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped. Exiting...")
}

var (
	// Актуальные значения переменных устанавливаются при сборке приложения.
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// buildInfo - возвращает информацию о сборке.
func buildInfo() string {
	return "Build version: " + buildVersion + "\n" +
		"Build date: " + buildDate + "\n" +
		"Build commit: " + buildCommit + "\n"
}
