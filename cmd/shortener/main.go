package main

import (
	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := app.NewConfig(4096, "http://localhost:8080/")
	a := app.NewApp(cfg, storage.NewMemory())
	h := handlers.NewShortener(a)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: h,
	}
	log.Fatal(server.ListenAndServe())
}
