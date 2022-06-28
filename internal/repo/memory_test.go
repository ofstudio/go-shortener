package repo

import (
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/stretchr/testify/suite"
	"testing"
)

type memoryRepoSuite struct {
	suite.Suite
	repo *MemoryRepo
}

func (suite *memoryRepoSuite) SetupTest() {
	suite.repo = NewMemoryRepo()
}

func (suite *memoryRepoSuite) TestUserCreate() {
	// Создаем первого пользователя
	user1 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user1))
	// Проверяем, что пользователю установлен ID=1
	suite.Equal(1, int(user1.ID))

	// Создаем второго пользователя
	user2 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user2))
	// Проверяем, что пользователю установлен ID больший, чем у первого
	suite.Greater(user2.ID, user1.ID)

	// Пытаемся создать пользователя с ID уже существующего пользователя
	user3 := &models.User{ID: user1.ID}
	suite.Equal(ErrDuplicate, suite.repo.UserCreate(user3))

	// Пытаемся создать пользователя из nil-объекта
	suite.Equal(ErrInvalidModel, suite.repo.UserCreate(nil))
}

func (suite *memoryRepoSuite) TestUserGetByID() {
	// Создаем пользователя
	user1 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user1))
	// Получаем пользователя по ID
	user2, err := suite.repo.UserGetByID(user1.ID)
	suite.NoError(err)
	// Проверяем, что пользователь совпадает с первым
	suite.Equal(user1, user2)

	// Пытаемся получить пользователя по несуществующему ID
	_, err = suite.repo.UserGetByID(user1.ID + 1)
	suite.Equal(ErrNotFound, err)
}

func (suite *memoryRepoSuite) TestShortURLCreate() {
	// Создаем первую сокращенную ссылку
	shortURL1 := &models.ShortURL{
		ID:          "12345",
		OriginalURL: "https://www.google.com",
		UserID:      1,
	}
	suite.NoError(suite.repo.ShortURLCreate(shortURL1))

	// Пытаемся создать вторую сокращенную ссылку с таким же ID
	shortURL2 := &models.ShortURL{
		ID:          "12345",
		OriginalURL: "https://www.baidu.com",
	}
	suite.Equal(ErrDuplicate, suite.repo.ShortURLCreate(shortURL2))

	// Пытаемся создать сокращенную ссылку из nil-объекта
	suite.Equal(ErrInvalidModel, suite.repo.ShortURLCreate(nil))
}

func (suite *memoryRepoSuite) TestShortURLGetById() {
	// Создаем первую сокращенную ссылку
	shortURL1 := &models.ShortURL{
		ID:          "12345",
		OriginalURL: "https://www.google.com",
		UserID:      1,
	}
	suite.NoError(suite.repo.ShortURLCreate(shortURL1))
	// Получаем сокращенную ссылку по ID
	shortURL2, err := suite.repo.ShortURLGetByID("12345")
	suite.NoError(err)
	// Проверяем, что сокращенная ссылка совпадает с первой
	suite.Equal(shortURL1, shortURL2, "should return the same shortURL")

	// Пытаемся получить сокращенную ссылку по несуществующему ID
	_, err = suite.repo.ShortURLGetByID("not-exist")
	suite.Equal(ErrNotFound, err)
}

func (suite *memoryRepoSuite) TestShortURLGetByUserId() {
	// Создаем первую сокращенную ссылку
	shortURL1 := &models.ShortURL{
		ID:          "12345",
		OriginalURL: "https://www.google.com",
		UserID:      1,
	}
	suite.NoError(suite.repo.ShortURLCreate(shortURL1))

	// Создаем вторую сокращенную ссылку с таким же UserID
	shortURL2 := &models.ShortURL{
		ID:          "67890",
		OriginalURL: "https://www.baidu.com",
		UserID:      1,
	}
	suite.NoError(suite.repo.ShortURLCreate(shortURL2))

	// Получаем сокращенные ссылки пользователя
	shortURLs, err := suite.repo.ShortURLGetByUserID(1)
	suite.NoError(err)
	// Проверяем, что получили такие же 2 сокращенные ссылки
	suite.Equal([]models.ShortURL{*shortURL1, *shortURL2}, shortURLs)

	// Пытаемся получить сокращенные ссылки пользователя по несуществующему ID
	_, err = suite.repo.ShortURLGetByUserID(2)
	suite.Equal(ErrNotFound, err)
}

func (suite *memoryRepoSuite) Test_autoIncrement() {
	// Создаем первого пользователя
	user1 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user1))
	// Проверяем, что первый пользователь имеет ID = 1
	suite.Equal(1, int(user1.ID))

	// Создаем второго пользователя
	user2 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user2))
	// Проверяем, что второй пользователь имеет ID = 2
	suite.Equal(2, int(user2.ID))

	// Создаем пользователя с ID=100
	user100 := &models.User{ID: 100}
	suite.NoError(suite.repo.UserCreate(user100))
	// Проверяем, что следующий пользователь должен быть с ID = 101
	user101 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user101))
	suite.Equal(101, int(user101.ID))

	// Создаем пользователя с ID меньшим, чем последний
	user90 := &models.User{ID: 90}
	suite.NoError(suite.repo.UserCreate(user90))
	// Следующий пользователь должен с ID = 102
	user102 := &models.User{}
	suite.NoError(suite.repo.UserCreate(user102))
	suite.Equal(102, int(user102.ID))
}

func TestMemoryRepoSuite(t *testing.T) {
	suite.Run(t, new(memoryRepoSuite))
}
