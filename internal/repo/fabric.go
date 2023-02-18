package repo

import (
	"github.com/rs/zerolog/log"

	"github.com/ofstudio/go-shortener/internal/config"
)

// Fabric - фабрика для создания репозитория.
// Если задан DatabaseDSN - используем SQL-репозиторий.
// Иначе, если задан fileStoragePath — используем AOF-репозиторий.
// Иначе используем репозиторий в памяти.
func Fabric(cfg *config.Config) (IRepo, error) {
	switch {
	case cfg.DatabaseDSN != "":
		log.Info().Msg("Using Postgres storage")
		return NewSQLRepo(cfg.DatabaseDSN)
	case cfg.FileStoragePath != "":
		log.Info().Msg("Using file storage")
		return NewAOFRepo(cfg.FileStoragePath)
	default:
		log.Info().Msg("Using in-memory storage")
		return NewMemoryRepo(), nil
	}
}
