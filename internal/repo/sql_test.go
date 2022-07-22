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
	repo          *SQLRepo
	testShortURLs []*models.ShortURL
}

func (suite *sqlRepoSuite) SetupTest() {
	var err error
	suite.repo, err = NewSQLRepo(dsn)
	suite.NoError(err)
	suite.NotNil(suite.repo)
	suite.testShortURLs = []*models.ShortURL{
		{ID: "12345", OriginalURL: "https://www.google.com", UserID: 1},
		{ID: "67890", OriginalURL: "https://www.baidu.com", UserID: 1},
		{ID: "aaaaa", OriginalURL: "https://www.qq.com", UserID: 1},
		{ID: "bbbbb", OriginalURL: "https://www.taobao.com", UserID: 2},
	}

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
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	actual, err := suite.repo.UserGetByID(context.Background(), user.ID)
	suite.NoError(err)
	suite.Equal(user.ID, actual.ID)
}

func (suite *sqlRepoSuite) TestUserGetByID_NotFound() {
	_, err := suite.repo.UserGetByID(context.Background(), 1)
	suite.Equal(ErrNotFound, err)
}

func (suite *sqlRepoSuite) TestShortURLCreate() {
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
}

func (suite *sqlRepoSuite) TestShortURLCreate_Duplicate() {
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	err := suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0])
	suite.Equal(ErrDuplicate, err)
}

func (suite *sqlRepoSuite) TestShortURLGetByID() {
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	actual, err := suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.Equal(suite.testShortURLs[0], actual)
}

func (suite *sqlRepoSuite) TestShortURLGetByID_NotFound() {
	_, err := suite.repo.ShortURLGetByID(context.Background(), "aaa")
	suite.Equal(ErrNotFound, err)
}

func (suite *sqlRepoSuite) TestShortURLGetByUserID() {
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[1]))
	shortURLs, err := suite.repo.ShortURLGetByUserID(context.Background(), suite.testShortURLs[0].UserID)
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
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	actual, err := suite.repo.ShortURLGetByOriginalURL(context.Background(), suite.testShortURLs[0].OriginalURL)
	suite.NoError(err)
	suite.Equal(suite.testShortURLs[0], actual)
}

func (suite *sqlRepoSuite) TestShortURLGetByURL_NotFound() {
	_, err := suite.repo.ShortURLGetByOriginalURL(context.Background(), "https://example.com")
	suite.Equal(ErrNotFound, err)
}

func (suite *sqlRepoSuite) TestShortURLDelete() {
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[1]))

	// Помечаем ссылку как удаленную
	suite.NoError(suite.repo.ShortURLDelete(context.Background(), suite.testShortURLs[0].UserID, suite.testShortURLs[0].ID))
	// Проверяем, что она помечена как удаленная
	actual, err := suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.Equal(true, actual.Deleted)

	// Пробуем пометить как удаленную ссылку другого пользователя
	suite.Equal(ErrNotFound, suite.repo.ShortURLDelete(context.Background(), 9999, suite.testShortURLs[1].ID))
	// Проверяем, что она не помечена как удаленная
	actual, err = suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[1].ID)
	suite.NoError(err)
	suite.Equal(false, actual.Deleted)
}

func (suite *sqlRepoSuite) TestShortURLDeleteBatch() {
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.UserCreate(context.Background(), &models.User{}))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[1]))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[2]))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[3]))

	// Пытаемся пометить ссылки как удаленные
	chA, chB := make(chan string), make(chan string)
	go func() {
		chA <- suite.testShortURLs[0].ID
		chA <- suite.testShortURLs[1].ID
		chB <- suite.testShortURLs[2].ID
		chB <- suite.testShortURLs[3].ID // <- Эта ссылка не будет удалена
		close(chA)
		close(chB)
	}()

	num, err := suite.repo.ShortURLDeleteBatch(context.Background(), 1, chA, chB)
	suite.NoError(err)
	suite.Equal(3, int(num))

	// Проверяем, что нужные ссылки были помечены как удаленные
	actual, err := suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[1].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[2].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[3].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(false, actual.Deleted, "should not be deleted")
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
