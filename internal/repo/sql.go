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
	st statements
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
	if r.st, err = prepareStmts(db); err != nil {
		return nil, err
	}
	return r, nil
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
	err := r.st[stmtUserCreate].QueryRowContext(ctx).Scan(&user.ID)
	return err
}

// UserGetByID - возвращает пользователя по его id.
func (r *SQLRepo) UserGetByID(ctx context.Context, id uint) (*models.User, error) {
	if r.db == nil {
		return nil, ErrDBNotInitialized
	}
	rows, err := r.st[stmtUserGetByID].QueryContext(ctx, id)
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
	_, err := r.st[stmtShortURLCreate].ExecContext(ctx, url.ID, url.OriginalURL, url.UserID)

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
	rows, err := r.st[stmtShortURLGetByID].QueryContext(ctx, id)
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
	rows, err := r.st[stmtShortURLGetByUserID].QueryContext(ctx, id)
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
	rows, err := r.st[stmtShortURLGetByOriginalURL].QueryContext(ctx, s)
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
