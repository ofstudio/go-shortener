package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.NewFromEnvAndCLI()
	if err != nil {
		log.Fatal(err)
	}

	var db storage.Interface
	// Если задан cfg.FileStoragePath, то используем файловый сторадж, иначе храним в памяти
	if cfg.FileStoragePath != "" {
		log.Println("Using file storage:", cfg.FileStoragePath)
		db, err = storage.NewAOFStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Using in-memory storage")
		db = storage.NewMemoryStorage()
	}

	srv := services.NewShortenerService(cfg, db)
	appHandlers := handlers.NewShortenerHandlers(srv)
	apiHandlers := handlers.NewAPIHandlers(srv)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", appHandlers.Routes())
	r.Mount("/api/", apiHandlers.Routes())

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	log.Printf("Starting shortener server at %s", cfg.ServerAddress)
	log.Fatal(server.ListenAndServe())
}
