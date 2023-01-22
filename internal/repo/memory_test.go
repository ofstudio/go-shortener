package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/go-shortener/internal/models"
)

type memoryRepoSuite struct {
	suite.Suite
	repo          *MemoryRepo
	testShortURLs []*models.ShortURL
}

func (suite *memoryRepoSuite) SetupTest() {
	suite.repo = NewMemoryRepo()
	suite.testShortURLs = []*models.ShortURL{
		{ID: "12345", OriginalURL: "https://www.google.com", UserID: 1},
		{ID: "67890", OriginalURL: "https://www.baidu.com", UserID: 1},
		{ID: "aaaaa", OriginalURL: "https://www.qq.com", UserID: 1},
		{ID: "bbbbb", OriginalURL: "https://www.taobao.com", UserID: 2},
	}
}

func (suite *memoryRepoSuite) TestUserCreate() {
	// Создаем первого пользователя
	user1 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user1))
	// Проверяем, что пользователю установлен ID=1
	suite.Equal(1, int(user1.ID))

	// Создаем второго пользователя
	user2 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user2))
	// Проверяем, что пользователю установлен ID больший, чем у первого
	suite.Greater(user2.ID, user1.ID)

	// Пытаемся создать пользователя с ID уже существующего пользователя
	user3 := &models.User{ID: user1.ID}
	suite.Equal(ErrDuplicate, suite.repo.UserCreate(context.Background(), user3))

	// Пытаемся создать пользователя из nil-объекта
	suite.Equal(ErrInvalidModel, suite.repo.UserCreate(context.Background(), nil))
}

func (suite *memoryRepoSuite) TestUserGetByID() {
	// Создаем пользователя
	user := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user))
	// Получаем пользователя по ID
	actual, err := suite.repo.UserGetByID(context.Background(), user.ID)
	suite.NoError(err)
	// Проверяем, что пользователь совпадает с первым
	suite.Equal(user, actual)

	// Пытаемся получить пользователя по несуществующему ID
	_, err = suite.repo.UserGetByID(context.Background(), user.ID+1)
	suite.Equal(ErrNotFound, err)
}

func (suite *memoryRepoSuite) TestShortURLCreate() {
	// Создаем первую сокращенную ссылку
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))

	// Пытаемся создать вторую сокращенную ссылку с таким же ID
	withDuplicateID := &models.ShortURL{
		ID:          suite.testShortURLs[0].ID,
		OriginalURL: "https://www.never.com",
		UserID:      1,
	}
	suite.Equal(ErrDuplicate, suite.repo.ShortURLCreate(context.Background(), withDuplicateID))

	// Пытаемся создать сокращенную ссылку из nil-объекта
	suite.Equal(ErrInvalidModel, suite.repo.ShortURLCreate(context.Background(), nil))
}

func (suite *memoryRepoSuite) TestShortURLGetById() {
	// Создаем первую сокращенную ссылку
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	// Получаем сокращенную ссылку по ID
	actual, err := suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	// Проверяем, что сокращенная ссылка совпадает с первой
	suite.Equal(suite.testShortURLs[0], actual, "should return the same actual")

	// Пытаемся получить сокращенную ссылку по несуществующему ID
	_, err = suite.repo.ShortURLGetByID(context.Background(), "not-exist")
	suite.Equal(ErrNotFound, err)
}

func (suite *memoryRepoSuite) TestShortURLGetByUserId() {
	// Создаем первую сокращенную ссылку
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))

	// Создаем вторую сокращенную ссылку с таким же UserID
	withSameUserID := &models.ShortURL{
		ID:          "67890",
		OriginalURL: "https://www.baidu.com",
		UserID:      suite.testShortURLs[0].UserID,
	}
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), withSameUserID))

	// Получаем сокращенные ссылки пользователя
	actual, err := suite.repo.ShortURLGetByUserID(context.Background(), 1)
	suite.NoError(err)
	// Проверяем, что получили такие же 2 сокращенные ссылки
	suite.Equal([]models.ShortURL{*suite.testShortURLs[0], *withSameUserID}, actual)

	// Пытаемся получить сокращенные ссылки пользователя по несуществующему ID
	actual, err = suite.repo.ShortURLGetByUserID(context.Background(), 2)
	suite.NoError(err)
	suite.Nil(actual)
}

func (suite *memoryRepoSuite) TestShortURLDelete() {
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(suite.repo.ShortURLCreate(context.Background(), suite.testShortURLs[1]))

	// Удаляем сокращенную ссылку
	suite.NoError(suite.repo.ShortURLDelete(context.Background(), suite.testShortURLs[0].UserID, suite.testShortURLs[0].ID))
	actual, err := suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.Equal(true, actual.Deleted, "should be deleted")

	// Пытаемся удалить ссылку несуществующего пользователя
	suite.Equal(ErrNotFound, suite.repo.ShortURLDelete(context.Background(), 9999, suite.testShortURLs[1].ID))
	actual, err = suite.repo.ShortURLGetByID(context.Background(), suite.testShortURLs[1].ID)
	suite.NoError(err)
	suite.Equal(false, actual.Deleted, "should not be deleted")

}

func (suite *memoryRepoSuite) TestShortURLDeleteBatch() {
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

func (suite *memoryRepoSuite) Test_autoIncrement() {
	// Создаем первого пользователя
	user1 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user1))
	// Проверяем, что первый пользователь имеет ID = 1
	suite.Equal(1, int(user1.ID))

	// Создаем второго пользователя
	user2 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user2))
	// Проверяем, что второй пользователь имеет ID = 2
	suite.Equal(2, int(user2.ID))

	// Создаем пользователя с ID=100
	user100 := &models.User{ID: 100}
	suite.NoError(suite.repo.UserCreate(context.Background(), user100))
	// Проверяем, что следующий пользователь должен быть с ID = 101
	user101 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user101))
	suite.Equal(101, int(user101.ID))

	// Создаем пользователя с ID меньшим, чем последний
	user90 := &models.User{ID: 90}
	suite.NoError(suite.repo.UserCreate(context.Background(), user90))
	// Следующий пользователь должен с ID = 102
	user102 := &models.User{}
	suite.NoError(suite.repo.UserCreate(context.Background(), user102))
	suite.Equal(102, int(user102.ID))
}

func TestMemoryRepoSuite(t *testing.T) {
	suite.Run(t, new(memoryRepoSuite))
}

// ======================
// Бенчмарки
// ======================

func BenchmarkMemoryRepo_UserGetByID(b *testing.B) {
	repo := NewMemoryRepo()
	user := &models.User{}
	ctx := context.Background()
	_ = repo.UserCreate(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.UserGetByID(ctx, user.ID)
	}
}

func BenchmarkMemoryRepo_ShortURLGetByID(b *testing.B) {
	repo := NewMemoryRepo()
	shortURL := &models.ShortURL{
		ID:          "1234",
		OriginalURL: "https://google.com",
		UserID:      1,
		Deleted:     false,
	}
	ctx := context.Background()
	_ = repo.ShortURLCreate(ctx, shortURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ShortURLGetByID(ctx, shortURL.ID)
	}
}

func BenchmarkMemoryRepo_ShortURLGetByUserID(b *testing.B) {
	repo := NewMemoryRepo()
	shortURL1 := &models.ShortURL{
		ID:          "1234",
		OriginalURL: "https://google.com",
		UserID:      1,
		Deleted:     false,
	}
	shortURL2 := &models.ShortURL{
		ID:          "5678",
		OriginalURL: "https://apple.com",
		UserID:      1,
		Deleted:     false,
	}
	ctx := context.Background()
	_ = repo.ShortURLCreate(ctx, shortURL1)
	_ = repo.ShortURLCreate(ctx, shortURL2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ShortURLGetByUserID(ctx, 1)
	}
}
