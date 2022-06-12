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
	cfg, err := config.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db := storage.NewMemoryStorage()
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
