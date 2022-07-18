package repo

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/stretchr/testify/suite"
	"testing"
)

const dsn = "postgres://autotest:autotest@localhost:5432/autotest"

func TestSQLRepoSuite(t *testing.T) {
	// Если тестовая БД не запущена - пропускаем тест
	if !testIsDBAvailable(dsn) {
		t.Skip("Database is not available: skipping SQLRepo suite")
	}
	suite.Run(t, new(sqlRepoSuite))
}

type sqlRepoSuite struct {
	suite.Suite
	repo *SQLRepo
}

func (suite *sqlRepoSuite) SetupTest() {
	var err error
	suite.repo, err = NewSQLRepo(dsn)
	suite.NoError(err)
	suite.NotNil(suite.repo)
}

func (suite *sqlRepoSuite) TearDownTest() {
	db := suite.repo.DB()
	suite.NotNil(db)
	// Удаляем таблицы
	_, err := db.Exec(`DROP TABLE IF EXISTS short_urls`)
	suite.NoError(err)
	_, err = db.Exec(`DROP TABLE IF EXISTS users`)
	suite.NoError(err)
	// Закрываем соединение
	suite.NoError(suite.repo.Close())
}

func (suite *sqlRepoSuite) TestUserCreate() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	suite.Equal(1, int(user.ID))
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	suite.Equal(2, int(user.ID))
}

func (suite *sqlRepoSuite) TestUserGetByID() {
	user1 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user1))
	user2, err := suite.repo.UserGetByID(context.Background(), user1.ID)
	suite.NoError(err)
	suite.Equal(user1.ID, user2.ID)
}

func (suite *sqlRepoSuite) TestUserGetByID_NotFound() {
	_, err := suite.repo.UserGetByID(context.Background(), 1)
	suite.Equal(ErrNotFound, err)
}

func (suite *sqlRepoSuite) TestShortURLCreate() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	shortURL := &models.ShortURL{ID: "aaa", OriginalURL: "https://example.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL))
}

func (suite *sqlRepoSuite) TestShortURLCreate_Duplicate() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	shortURL := &models.ShortURL{ID: "aaa", OriginalURL: "https://example.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL))
	err := suite.repo.ShortURLCreate(context.Background(), shortURL)
	suite.Equal(ErrDuplicate, err)
}

func (suite *sqlRepoSuite) TestShortURLGetByID() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	shortURL := &models.ShortURL{ID: "aaa", OriginalURL: "https://example.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL))
	shortURL2, err := suite.repo.ShortURLGetByID(context.Background(), shortURL.ID)
	suite.NoError(err)
	suite.Equal(shortURL, shortURL2)
}

func (suite *sqlRepoSuite) TestShortURLGetByID_NotFound() {
	_, err := suite.repo.ShortURLGetByID(context.Background(), "aaa")
	suite.Equal(ErrNotFound, err)
}

func (suite *sqlRepoSuite) TestShortURLGetByUserID() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	shortURL := &models.ShortURL{ID: "aaa", OriginalURL: "https://example.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL))
	shortURL2 := &models.ShortURL{ID: "bbb", OriginalURL: "https://another.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL2))
	shortURLs, err := suite.repo.ShortURLGetByUserID(context.Background(), user.ID)
	suite.NoError(err)
	suite.Equal(2, len(shortURLs))
}

func (suite *sqlRepoSuite) TestShortURLGetByUserID_UserNotFound() {
	urls, err := suite.repo.ShortURLGetByUserID(context.Background(), 1)
	suite.NoError(err)
	suite.Nil(urls)
}

func (suite *sqlRepoSuite) TestShortURLGetByUserID_NoURLs() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	urls, err := suite.repo.ShortURLGetByUserID(context.Background(), user.ID)
	suite.NoError(err)
	suite.Nil(urls)
}

func (suite *sqlRepoSuite) TestShortURLGetByURL() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	shortURL := &models.ShortURL{ID: "aaa", OriginalURL: "https://example.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL))
	shortURL2, err := suite.repo.ShortURLGetByOriginalURL(context.Background(), shortURL.OriginalURL)
	suite.NoError(err)
	suite.Equal(shortURL, shortURL2)
}

func (suite *sqlRepoSuite) TestShortURLGetByURL_NotFound() {
	_, err := suite.repo.ShortURLGetByOriginalURL(context.Background(), "https://example.com")
	suite.Equal(ErrNotFound, err)
}

func (suite *sqlRepoSuite) TestShortURLDeleteBatch() {
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	shortURL := &models.ShortURL{ID: "aaa", OriginalURL: "https://example.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL))
	shortURL2 := &models.ShortURL{ID: "bbb", OriginalURL: "https://another.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL2))
	shortURL3 := &models.ShortURL{ID: "ccc", OriginalURL: "https://more.com", UserID: user.ID}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), shortURL3))

	// Удаляем ссылки пользователя
	n, err := suite.repo.ShortURLDeleteBatch(context.Background(), user.ID, []string{"aaa", "bbb"})
	suite.NoError(err)
	suite.Equal(2, int(n))
	shortURLs, err := suite.repo.ShortURLGetByUserID(context.Background(), user.ID)
	suite.NoError(err)
	suite.Equal(1, len(shortURLs))

	// Удаляем ссылки не принадлежащие пользователю
	n, err = suite.repo.ShortURLDeleteBatch(context.Background(), 1000, []string{"ccc"})
	suite.NoError(err)
	suite.Equal(0, int(n))
	shortURLs, err = suite.repo.ShortURLGetByUserID(context.Background(), user.ID)
	suite.NoError(err)
	suite.Equal(1, len(shortURLs))
}

func testIsDBAvailable(dsn string) bool {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return false
	}
	//goland:noinspection GoUnhandledErrorResult
	defer db.Close()
	return db.Ping() == nil
}
