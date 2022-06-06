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
	cfg := &config.Config{
		UrlMaxLen: 4096,
		PublicURL: "http://localhost:8080/",
	}
	db := storage.NewMemoryStorage()
	srv := services.NewShortenerService(cfg, db)
	h := handlers.NewShortenerHandlers(srv)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", h.Routes())

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: r,
	}
	log.Fatal(server.ListenAndServe())
}
