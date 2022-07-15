package repo

import (
	"context"
	"database/sql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ofstudio/go-shortener/internal/models"
	"log"
)

type SQLRepo struct {
	db *sql.DB
	// Заранее подготовленные запросы к БД
	userCreateStmt          *sql.Stmt
	userGetByIDStmt         *sql.Stmt
	shortURLCreateStmt      *sql.Stmt
	shortURLGetByIDStmt     *sql.Stmt
	shortURLGetByUserIDStmt *sql.Stmt
	shortURLGetByURLStmt    *sql.Stmt
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
	if err = r.prepareStatements(); err != nil {
		return nil, err
	}
	return r, nil
}

// prepareStatements - подготавливает запросы к БД
func (r *SQLRepo) prepareStatements() error {
	if r.db == nil {
		return ErrDBNotInitialized
	}

	var err error

	r.userCreateStmt, err = r.db.Prepare(`
		INSERT INTO users (id)
		VALUES (DEFAULT)
		RETURNING id
	`)
	if err != nil {
		return err
	}

	r.userGetByIDStmt, err = r.db.Prepare(`
		SELECT id FROM users 
	  	WHERE id = $1
	`)
	if err != nil {
		return err
	}

	r.shortURLCreateStmt, err = r.db.Prepare(`
		INSERT INTO short_urls (id, original_url, user_id)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return err
	}

	r.shortURLGetByIDStmt, err = r.db.Prepare(`
		SELECT id, original_url, user_id FROM short_urls 
		WHERE id = $1
	`)
	if err != nil {
		return err
	}

	r.shortURLGetByUserIDStmt, err = r.db.Prepare(`
		SELECT id, original_url, user_id FROM short_urls 
		WHERE user_id = $1
	`)
	if err != nil {
		return err
	}

	r.shortURLGetByURLStmt, err = r.db.Prepare(`
		SELECT id, original_url, user_id FROM short_urls
		WHERE original_url = $1
	`)
	if err != nil {
		return err
	}

	return nil
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
	if r.db == nil {
		return ErrDBNotInitialized
	}
	err := r.userCreateStmt.QueryRowContext(ctx).Scan(&user.ID)
	return err
}

// UserGetByID - возвращает пользователя по его id.
func (r *SQLRepo) UserGetByID(ctx context.Context, id uint) (*models.User, error) {
	if r.db == nil {
		return nil, ErrDBNotInitialized
	}
	rows, err := r.userGetByIDStmt.QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrNotFound
	}
	var u models.User
	if err = rows.Scan(&u.ID); err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &u, nil
}

// ShortURLCreate - добавляет новую сокращенную ссылку в репозиторий.
func (r *SQLRepo) ShortURLCreate(ctx context.Context, url *models.ShortURL) error {
	if r.db == nil {
		return ErrDBNotInitialized
	}
	_, err := r.shortURLCreateStmt.ExecContext(ctx, url.ID, url.OriginalURL, url.UserID)

	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return ErrDuplicate
	}
	return err
}

// ShortURLGetByID - возвращает сокращенную ссылку по ее id.
func (r *SQLRepo) ShortURLGetByID(ctx context.Context, id string) (*models.ShortURL, error) {
	if r.db == nil {
		return nil, ErrDBNotInitialized
	}
	rows, err := r.shortURLGetByIDStmt.QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrNotFound
	}
	var u models.ShortURL
	if err = rows.Scan(&u.ID, &u.OriginalURL, &u.UserID); err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &u, nil
}

// ShortURLGetByUserID - возвращает сокращенные ссылки пользователя.
// Если пользователь не найден, или у пользователя нет ссылок возвращает nil.
func (r *SQLRepo) ShortURLGetByUserID(ctx context.Context, id uint) ([]models.ShortURL, error) {
	if r.db == nil {
		return nil, ErrDBNotInitialized
	}
	rows, err := r.shortURLGetByUserIDStmt.QueryContext(ctx, id)
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
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return urls, nil
}

// ShortURLGetByOriginalURL - возвращает сокращенную ссылку по ее оригинальному url.
func (r *SQLRepo) ShortURLGetByOriginalURL(ctx context.Context, s string) (*models.ShortURL, error) {
	if r.db == nil {
		return nil, ErrDBNotInitialized
	}
	rows, err := r.shortURLGetByURLStmt.QueryContext(ctx, s)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotFound
	}
	var u models.ShortURL
	if err = rows.Scan(&u.ID, &u.OriginalURL, &u.UserID); err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &u, nil
}
