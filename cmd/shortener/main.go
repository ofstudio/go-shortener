package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := app.NewConfig(4096, "http://localhost:8080/")
	a := app.NewApp(cfg, storage.NewMemory())

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/", handlers.NewURLsResource(a).Routes())

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: r,
	}
	log.Fatal(server.ListenAndServe())
}
