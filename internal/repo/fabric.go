package repo

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"log"
)

// Fabric - фабрика для создания репозитория.
// Если задан DatabaseDSN - используем SQL-репозиторий.
// Иначе, если задан fileStoragePath — используем AOF-репозиторий.
// Иначе используем репозиторий в памяти.
func Fabric(cfg *config.Config) Repo {
	switch {
	case cfg.DatabaseDSN != "":
		log.Print("Using Postgres storage")
		return MustNewSQLRepo(cfg.DatabaseDSN)
	case cfg.FileStoragePath != "":
		log.Print("Using file storage")
		return MustNewAOFRepo(cfg.FileStoragePath)
	default:
		log.Print("Using in-memory storage")
		return NewMemoryRepo()
	}
}
