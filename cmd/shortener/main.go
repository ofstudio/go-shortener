package main

import (
	"context"
	"os"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/config"
)

var (
	// Актуальные значения переменных устанавливаются при сборке приложения.
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	writer := zerolog.NewConsoleWriter()
	writer.TimeFormat = time.RFC3339
	log.Logger = zerolog.New(writer).
		Level(zerolog.InfoLevel).
		With().Timestamp().
		Logger()
}

func main() {

	// Выводим информацию о сборке
	log.Info().
		Str("commit", buildCommit).
		Str("date", buildDate).
		Str("version", buildVersion).
		Msg("Build info")

	// Считываем конфигурацию
	cfg, err := config.Compose(
		config.Default,                     // Значения по умолчанию
		config.FromJSONFile(os.Args[:]...), // Значения из JSON-файла
		config.FromEnv,                     // Значения из переменных окружения
		config.FromCLI(os.Args[1:]...),     // Значения из флагов командной строки
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while loading config")
	}

	// Контекст для остановки приложения
	ctx, cancel := app.ContextWithShutdown(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Создаем и запускаем приложение
	a := app.NewApp(cfg)
	if err = a.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Application fatal error")
	}

	log.Info().Msg("Exiting")
}
