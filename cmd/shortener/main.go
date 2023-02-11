package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/server"
	"github.com/ofstudio/go-shortener/pkg/shutdown"
)

func main() {

	// Выводим информацию о сборке
	fmt.Print(buildInfo())

	// Считываем конфигурацию
	cfg, err := config.Compose(
		config.Default,                      // Значения по умолчанию
		config.FromJSONFile(os.Args[1:]...), // Значения из JSON-файла
		config.FromEnv,                      // Значения из переменных окружения
		config.FromCLI(os.Args[1:]...),      // Значения из флагов командной строки
	)
	if err != nil {
		log.Fatal(err)
	}

	// Контекст для остановки приложения
	ctx, cancel := shutdown.ContextWithShutdown(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	// Создаём сервер
	srv := server.NewServer(cfg)
	log.Printf("Starting server at %s", cfg.ServerAddress)
	if err = srv.Start(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Exiting...")
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
