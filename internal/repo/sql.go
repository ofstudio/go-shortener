package repo

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ofstudio/go-shortener/internal/models"
	"log"
)

type SQLRepo struct {
	db *sql.DB
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
	r := &SQLRepo{db: db}
	if err = r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

// migrate - создает таблицы в базе данных
func (r *SQLRepo) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY
		)
`)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS short_urls (
			id VARCHAR(127) PRIMARY KEY,
			original_url VARCHAR(4096) NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		)
	`)

	return err
}

// DB - возвращает подключение к базе данных
func (r *SQLRepo) DB() *sql.DB {
	return r.db
}

// Close - закрывает подключение к базе данных
func (r *SQLRepo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// UserCreate - добавляет нового пользователя в репозиторий.
func (r *SQLRepo) UserCreate(ctx context.Context, user *models.User) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (id)
		VALUES (DEFAULT)
		RETURNING id
	`).Scan(&user.ID)
	return err
}

// UserGetByID - возвращает пользователя по его id.
func (r *SQLRepo) UserGetByID(ctx context.Context, id uint) (*models.User, error) {
	row, err := r.db.QueryContext(ctx, `
		SELECT id FROM users 
	  	WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer row.Close()
	if !row.Next() {
		return nil, ErrNotFound
	}
	var u models.User
	if err = row.Scan(&u.ID); err != nil {
		return nil, err
	}
	return &u, nil
}

// ShortURLCreate - добавляет новую сокращенную ссылку в репозиторий.
func (r *SQLRepo) ShortURLCreate(ctx context.Context, url *models.ShortURL) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO short_urls (id, original_url, user_id)
		VALUES ($1, $2, $3)
	`, url.ID, url.OriginalURL, url.UserID)
	if err != nil {
		return err
	}
	return nil
}

// ShortURLGetByID - возвращает сокращенную ссылку по ее id.
func (r *SQLRepo) ShortURLGetByID(ctx context.Context, id string) (*models.ShortURL, error) {
	row, err := r.db.QueryContext(ctx, `
		SELECT id, original_url, user_id FROM short_urls 
		WHERE id = $1
		`, id)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer row.Close()
	if !row.Next() {
		return nil, ErrNotFound
	}
	var u models.ShortURL
	if err = row.Scan(&u.ID, &u.OriginalURL, &u.UserID); err != nil {
		return nil, err
	}
	return &u, nil
}

// ShortURLGetByUserID - возвращает сокращенные ссылки пользователя.
// Если пользователь не найден, или у пользователя нет ссылок возвращает nil.
func (r *SQLRepo) ShortURLGetByUserID(ctx context.Context, id uint) ([]models.ShortURL, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, original_url, user_id FROM short_urls 
		WHERE user_id = $1
		`, id)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()
	var urls []models.ShortURL
	for rows.Next() {
		var u models.ShortURL
		if err = rows.Scan(&u.ID, &u.OriginalURL, &u.ID); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
