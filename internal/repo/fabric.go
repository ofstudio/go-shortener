package repo

// MustNewRepoFabric - фабрика для создания репозитория.
// Если задан fileStoragePath, то используем AOF-репозиторий.
// Иначе используем репозиторий в памяти.
func MustNewRepoFabric(fileStoragePath string) Repo {
	if fileStoragePath == "" {
		return NewMemoryRepo()
	}
	return MustNewAOFRepo(fileStoragePath)
}
