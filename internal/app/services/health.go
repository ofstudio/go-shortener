package services

import "database/sql"

// HealthService - healthcheck-сервис приложения
type HealthService struct {
	db *sql.DB
}

func NewHealthService(db *sql.DB) *HealthService {
	return &HealthService{db: db}
}

// Check - выполняет проверку приложения
func (s *HealthService) Check() error {
	if s.db == nil {
		return nil
	}
	return s.db.Ping()
}
