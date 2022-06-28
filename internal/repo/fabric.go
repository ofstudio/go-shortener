package repo

import "github.com/ofstudio/go-shortener/internal/app/config"

// Fabric - фабрика для создания репозитория.
// Если задан DatabaseDSN - используем SQL-репозиторий.
// Иначе, если задан fileStoragePath — используем AOF-репозиторий.
// Иначе используем репозиторий в памяти.
func Fabric(cfg *config.Config) Repo {
	switch {
	case cfg.DatabaseDSN != "":
		return MustNewSQLRepo(cfg.DatabaseDSN)
	case cfg.FileStoragePath != "":
		return MustNewAOFRepo(cfg.FileStoragePath)
	default:
		return NewMemoryRepo()
	}
}
