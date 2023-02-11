package repo

import (
	"context"
	"database/sql"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/ofstudio/go-shortener/internal/models"
)

// SQLRepo - реализация IRepo для хранения данных в PostgreSQL.
type SQLRepo struct {
	db *sql.DB
	st statements
}

// NewSQLRepo - конструктор репозитория SQLRepo.
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

// UserCount - возвращает количество пользователей в репозитории.
func (r *SQLRepo) UserCount(ctx context.Context) (int, error) {
	if r.db == nil {
		return 0, ErrDBNotInitialized
	}
	var count int
	err := r.st[stmtUserCount].QueryRowContext(ctx).Scan(&count)
	return count, err
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
	if err = rows.Scan(&u.ID, &u.OriginalURL, &u.UserID, &u.Deleted); err != nil {
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
		if err = rows.Scan(&u.ID, &u.OriginalURL, &u.UserID, &u.Deleted); err != nil {
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
	if err = rows.Scan(&u.ID, &u.OriginalURL, &u.UserID, &u.Deleted); err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return &u, nil
}

// ShortURLDelete - помечает удаленной короткую ссылку пользователя по ее id.
func (r *SQLRepo) ShortURLDelete(_ context.Context, userID uint, id string) error {
	if r.db == nil {
		return ErrDBNotInitialized
	}
	res, err := r.st[stmtShortURLDelete].ExecContext(context.Background(), userID, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return err
}

// ShortURLDeleteBatch - помечает удаленными сокращенных ссылок пользователя по их id.
// Принимает на вход список каналов для передачи идентификаторов.
// Возвращает количество помеченных удаленными ссылок.
func (r *SQLRepo) ShortURLDeleteBatch(ctx context.Context, userID uint, chans ...chan string) (int64, error) {
	if r.db == nil {
		return 0, ErrDBNotInitialized
	}

	// Мультиплексируем каналы chans в один канал ch.
	ch := fanIn(ctx, chans...)
	var ids []string
	// Читаем значения из канала и собираем все id для удаления один слайс
	for id := range ch {
		ids = append(ids, id)
	}

	// Удаляем ссылки по их id.
	if len(ids) == 0 {
		return 0, nil
	}
	res, err := r.st[stmtShortURLDeleteBatch].ExecContext(ctx, userID, ids)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ShortURLCount - возвращает количество сокращенных ссылок в репозитории.
func (r *SQLRepo) ShortURLCount(ctx context.Context) (int, error) {
	if r.db == nil {
		return 0, ErrDBNotInitialized
	}
	var count int
	err := r.st[stmtShortURLCount].QueryRowContext(ctx).Scan(&count)
	return count, err
}
