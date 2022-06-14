package main

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/pkg/middleware"
	"github.com/ofstudio/go-shortener/pkg/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Считываем конфигурацию.
	cfg, err := config.NewFromEnvAndCLI()
	if err != nil {
		log.Fatal(err)
	}

	// Создаём хранилище.
	// Если задан cfg.FileStoragePath, то используем файловый сторадж, иначе храним в памяти.
	var db storage.Interface
	if cfg.FileStoragePath != "" {
		log.Println("Using append-only file storage:", cfg.FileStoragePath)
		db, err = storage.NewAOFStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Using in-memory storage")
		db = storage.NewMemoryStorage()
	}

	// Создаём сервис и обработчики запросов.
	srv := services.NewShortenerService(cfg, db)
	appHandlers := handlers.NewShortenerHandlers(srv)
	apiHandlers := handlers.NewAPIHandlers(srv)

	// Создаём маршрутизатор.
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)

	// Middleware для декомпрессии запросов.
	r.Use(middleware.Decompressor)

	// Middleware для компрессии ответов.
	// Параметр minSize рекомендуется равным middleware.MTUSize.
	// Значение 0 означает сжатие ответов любой длины и используется в целях демонстрации.
	r.Use(middleware.NewCompressor(0, gzip.BestSpeed).AddType("application/json").Handler)

	// Добавляем маршруты для обработки запросов.
	r.Mount("/", appHandlers.Routes())
	r.Mount("/api/", apiHandlers.Routes())

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
		_ = server.Close()
	}()

	// Запускаем сервер.
	log.Printf("Starting http server at %s", cfg.ServerAddress)
	switch server.ListenAndServe() {
	case http.ErrServerClosed:
		log.Println("Server stopped. Exiting...")
	default:
		log.Fatal(err)
	}
}
