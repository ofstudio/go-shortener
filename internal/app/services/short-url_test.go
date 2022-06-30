package services

import (
	"context"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
)

type shortURLServiceSuite struct {
	cfg *config.Config
	*ShortURLService
	*UserService
	suite.Suite
}

func (suite *shortURLServiceSuite) SetupTest() {
	var err error
	os.Clearenv()
	suite.cfg, err = config.NewFromEnv()
	suite.Require().NoError(err)
	r := repo.NewMemoryRepo()
	suite.ShortURLService = NewShortURLService(suite.cfg, r)
	suite.UserService = NewUserService(suite.cfg, r)
	suite.Require().NoError(suite.UserService.Create(context.Background(), &models.User{}))
}

func (suite *shortURLServiceSuite) TestCreate() {
	// Успешное создание короткой ссылки
	suite.Run("success", func() {
		shortURL, err := suite.ShortURLService.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		suite.NotNil(shortURL)
		suite.Equal("https://google.com", shortURL.OriginalURL)
		suite.Equal(1, int(shortURL.UserID))
		suite.NotEmpty(shortURL.ID)
	})

	// Невалидный URL
	suite.Run("invalid url", func() {
		// Невалидный URL
		_, err := suite.ShortURLService.Create(context.Background(), 1, "invalid url")
		suite.Equal(ErrValidation, err)
		// Недопустимый протокол
		_, err = suite.ShortURLService.Create(context.Background(), 1, "file:///tmp/test.txt")
		suite.Equal(ErrValidation, err)
		// Пустой URL
		_, err = suite.ShortURLService.Create(context.Background(), 1, "")
		suite.Equal(ErrValidation, err)
		// Слишком длинный URL
		_, err = suite.ShortURLService.Create(context.Background(), 1, "https://google.com/"+strings.Repeat("a", models.URLMaxLen))
		suite.Equal(ErrValidation, err)
	})

	// Несуществующий пользователь
	suite.Run("invalid user", func() {
		_, err := suite.ShortURLService.Create(context.Background(), 100, "https://google.com")
		suite.Equal(ErrNotFound, err)
	})
}

func (suite *shortURLServiceSuite) TestGetByID() {
	// Успешное получение короткой ссылки
	suite.Run("success", func() {
		shortURL, err := suite.ShortURLService.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		suite.NotNil(shortURL)
		shortURL, err = suite.ShortURLService.GetByID(context.Background(), shortURL.ID)
		suite.NoError(err)
		suite.NotNil(shortURL)
		suite.Equal("https://google.com", shortURL.OriginalURL)
		suite.Equal(1, int(shortURL.UserID))
		suite.NotEmpty(shortURL.ID)
	})

	// Несуществующая короткая ссылка
	suite.Run("not found", func() {
		_, err := suite.ShortURLService.GetByID(context.Background(), "not found")
		suite.Equal(ErrNotFound, err)
	})
}

func (suite *shortURLServiceSuite) TestGetByUserID() {
	// Успешное получение коротких ссылок пользователя
	suite.Run("success", func() {
		_, err := suite.ShortURLService.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)
		_, err = suite.ShortURLService.Create(context.Background(), 1, "https://ya.ru")
		suite.NoError(err)
		shortURLs, err := suite.ShortURLService.GetByUserID(context.Background(), 1)
		suite.NoError(err)
		suite.NotNil(shortURLs)
		suite.Len(shortURLs, 2)
		suite.Equal("https://google.com", shortURLs[0].OriginalURL)
		suite.Equal("https://ya.ru", shortURLs[1].OriginalURL)
	})

	// У пользователя нет коротких ссылок
	suite.Run("no short urls", func() {
		user := &models.User{}
		suite.Require().NoError(suite.UserService.Create(context.Background(), user))
		shortURLs, err := suite.ShortURLService.GetByUserID(context.Background(), user.ID)
		suite.NoError(err)
		suite.Nil(shortURLs)
	})

	// Несуществующий пользователь
	suite.Run("invalid user", func() {
		urls, err := suite.ShortURLService.GetByUserID(context.Background(), 100)
		suite.NoError(err)
		suite.Nil(urls)
	})
}

func (suite *shortURLServiceSuite) TestResolve() {
	shortURL, err := suite.ShortURLService.Create(context.Background(), 1, "https://google.com")
	suite.NoError(err)
	suite.Equal(suite.cfg.BaseURL.String()+shortURL.ID, suite.ShortURLService.Resolve(shortURL.ID))
}

func TestShortURLServiceSuite(t *testing.T) {
	suite.Run(t, new(shortURLServiceSuite))
}
