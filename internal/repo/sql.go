package repo

import (
	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

type SQLRepo struct {
	db          *sql.DB
	*MemoryRepo // mock
}

func MustNewSQLRepo(dsn string) *SQLRepo {
	r, err := NewSQLRepo(dsn)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func NewSQLRepo(dsn string) (*SQLRepo, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &SQLRepo{db: db, MemoryRepo: NewMemoryRepo()}, err
}

func (r *SQLRepo) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY
		);
		CREATE TABLE IF NOT EXISTS short_urls (
			id VARCHAR(255) PRIMARY KEY,
			original_url VARCHAR(4096) NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		);
	`)
	return err
}

func (r *SQLRepo) DB() *sql.DB {
	return r.db
}

func (r *SQLRepo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
