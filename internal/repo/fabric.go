package repo

import (
	"log"

	"github.com/ofstudio/go-shortener/internal/app/config"
)

// Fabric - фабрика для создания репозитория.
// Если задан DatabaseDSN - используем SQL-репозиторий.
// Иначе, если задан fileStoragePath — используем AOF-репозиторий.
// Иначе используем репозиторий в памяти.
func Fabric(cfg *config.Config) (IRepo, error) {
	switch {
	case cfg.DatabaseDSN != "":
		log.Print("Using Postgres storage")
		return NewSQLRepo(cfg.DatabaseDSN)
	case cfg.FileStoragePath != "":
		log.Print("Using file storage")
		return NewAOFRepo(cfg.FileStoragePath)
	default:
		log.Print("Using in-memory storage")
		return NewMemoryRepo(), nil
	}
}
