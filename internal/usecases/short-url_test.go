package usecases

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
)

type shortURLSuite struct {
	cfg *config.Config
	*ShortURL
	*User
	suite.Suite
}

func (suite *shortURLSuite) SetupTest() {
	var err error
	os.Clearenv()
	suite.cfg, _ = config.Default(nil)
	suite.Require().NoError(err)
	r := repo.NewMemoryRepo()
	suite.ShortURL = NewShortURL(suite.cfg, r)
	suite.User = NewUser(suite.cfg, r)
	suite.Require().NoError(suite.User.Create(context.Background(), &models.User{}))
}

func (suite *shortURLSuite) TestCreate() {
	// Успешное создание короткой ссылки
	suite.Run("success", func() {
		shortURL, err := suite.ShortURL.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		suite.NotNil(shortURL)
		suite.Equal("https://google.com", shortURL.OriginalURL)
		suite.Equal(1, int(shortURL.UserID))
		suite.NotEmpty(shortURL.ID)
	})

	// Невалидный URL
	suite.Run("invalid url", func() {
		// Невалидный URL
		_, err := suite.ShortURL.Create(context.Background(), 1, "invalid url")
		suite.Equal(app.ErrValidation, err)
		// Недопустимый протокол
		_, err = suite.ShortURL.Create(context.Background(), 1, "file:///tmp/test.txt")
		suite.Equal(app.ErrValidation, err)
		// Пустой URL
		_, err = suite.ShortURL.Create(context.Background(), 1, "")
		suite.Equal(app.ErrValidation, err)
		// Слишком длинный URL
		_, err = suite.ShortURL.Create(context.Background(), 1, "https://google.com/"+strings.Repeat("a", models.URLMaxLen))
		suite.Equal(app.ErrValidation, err)
	})

	// Несуществующий пользователь
	suite.Run("invalid user", func() {
		_, err := suite.ShortURL.Create(context.Background(), 100, "https://google.com")
		suite.Equal(app.ErrNotFound, err)
	})

	suite.Run("duplicate url", func() {
		_, err := suite.ShortURL.Create(context.Background(), 1, "https://duplicate.com")
		suite.NoError(err)
		_, err = suite.ShortURL.Create(context.Background(), 1, "https://duplicate.com")
		suite.Equal(app.ErrDuplicate, err)
	})
}

func (suite *shortURLSuite) TestGetByID() {
	// Успешное получение короткой ссылки
	suite.Run("success", func() {
		shortURL, err := suite.ShortURL.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		suite.NotNil(shortURL)
		shortURL, err = suite.ShortURL.GetByID(context.Background(), shortURL.ID)
		suite.NoError(err)
		suite.NotNil(shortURL)
		suite.Equal("https://google.com", shortURL.OriginalURL)
		suite.Equal(1, int(shortURL.UserID))
		suite.NotEmpty(shortURL.ID)
	})

	// Несуществующая короткая ссылка
	suite.Run("not found", func() {
		_, err := suite.ShortURL.GetByID(context.Background(), "not found")
		suite.Equal(app.ErrNotFound, err)
	})
}

func (suite *shortURLSuite) TestGetByUserID() {
	// Успешное получение коротких ссылок пользователя
	suite.Run("success", func() {
		_, err := suite.ShortURL.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		_, err = suite.ShortURL.Create(context.Background(), 1, "https://ya.ru")
		suite.NoError(err)
		shortURLs, err := suite.ShortURL.GetByUserID(context.Background(), 1)
		suite.NoError(err)
		suite.NotNil(shortURLs)
		suite.Len(shortURLs, 2)
		suite.Equal("https://google.com", shortURLs[0].OriginalURL)
		suite.Equal("https://ya.ru", shortURLs[1].OriginalURL)
	})

	// У пользователя нет коротких ссылок
	suite.Run("no short urls", func() {
		user := &models.User{}
		suite.Require().NoError(suite.User.Create(context.Background(), user))
		shortURLs, err := suite.ShortURL.GetByUserID(context.Background(), user.ID)
		suite.NoError(err)
		suite.Nil(shortURLs)
	})

	// Несуществующий пользователь
	suite.Run("invalid user", func() {
		urls, err := suite.ShortURL.GetByUserID(context.Background(), 100)
		suite.NoError(err)
		suite.Nil(urls)
	})
}

func (suite *shortURLSuite) TestGetByOriginalURL() {
	// Успешное получение короткой ссылки по оригинальной
	suite.Run("success", func() {
		shortURL, err := suite.ShortURL.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		suite.NotNil(shortURL)
		shortURL, err = suite.ShortURL.GetByOriginalURL(context.Background(), "https://google.com")
		suite.NoError(err)
		suite.NotNil(shortURL)
		suite.Equal("https://google.com", shortURL.OriginalURL)
		suite.Equal(1, int(shortURL.UserID))
		suite.NotEmpty(shortURL.ID)
	})

	// Несуществующая оригинальная ссылка
	suite.Run("not found", func() {
		_, err := suite.ShortURL.GetByOriginalURL(context.Background(), "not found")
		suite.Equal(app.ErrNotFound, err)
	})
}

func (suite *shortURLSuite) TestResolve() {
	shortURL, err := suite.ShortURL.Create(context.Background(), 1, "https://google.com")
	suite.NoError(err)
	suite.Equal(suite.cfg.BaseURL.String()+shortURL.ID, suite.ShortURL.Resolve(shortURL.ID))
}

func TestShortURLSuite(t *testing.T) {
	suite.Run(t, new(shortURLSuite))
}
