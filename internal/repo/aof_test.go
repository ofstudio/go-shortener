package repo

import (
	"context"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
)

type aofRepoSuite struct {
	suite.Suite
	tmpDir        string
	filePath      string
	testShortURLs []*models.ShortURL
}

func (suite *aofRepoSuite) SetupTest() {
	var err error
	suite.tmpDir, err = os.MkdirTemp("", "aof_test-*")
	suite.NoError(err)
	suite.filePath = suite.tmpDir + "/shortener.aof"
	suite.testShortURLs = []*models.ShortURL{
		{ID: "12345", OriginalURL: "https://www.google.com", UserID: 1},
		{ID: "67890", OriginalURL: "https://www.baidu.com", UserID: 1},
		{ID: "aaaaa", OriginalURL: "https://www.qq.com", UserID: 1},
		{ID: "bbbbb", OriginalURL: "https://www.taobao.com", UserID: 2},
	}
}

func (suite *aofRepoSuite) TearDownTest() {
	suite.NoError(os.RemoveAll(suite.tmpDir))
}

func (suite *aofRepoSuite) TestNewAOFRepo() {
	// Создаем новый пустой репозиторий
	repo, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	suite.NotNil(repo)
	suite.NoError(repo.Close())

	// Пытаемся создать репозиторий из файла с некорректными данными
	suite.invalidJSONFile()
	_, err = NewAOFRepo(suite.filePath)
	suite.Equal(ErrAOFRead, err)

	// Пытаемся создать репозиторий из файла с некорректными JSON-структурами
	suite.invalidJSONStruct()
	_, err = NewAOFRepo(suite.filePath)
	suite.Equal(ErrAOFStructure, err)

	// Пытаемся создать репозиторий из несуществующего файла
	_, err = NewAOFRepo(suite.tmpDir + "/**/*")
	suite.Equal(ErrAOFOpen, err)

	// Пытаемся создать репозиторий из файла с одинаковыми ID пользователей
	suite.duplicateUser()
	_, err = NewAOFRepo(suite.filePath)
	suite.Equal(ErrDuplicate, err)

	// Пытаемся создать репозиторий из файла с одинаковыми ID сокращенных ссылок
	suite.duplicateShortURL()
	_, err = NewAOFRepo(suite.filePath)
	suite.Equal(ErrDuplicate, err)
}

func (suite *aofRepoSuite) TestAOFRepo_UserCreate() {
	// Создаем репозиторий и записываем в него пользователя
	repo1, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	suite.NoError(repo1.UserCreate(context.Background(), &models.User{ID: 100}))
	suite.NoError(repo1.Close())

	// Открываем репозиторий и проверяем, что пользователь записан в него
	repo2, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	actual, err := repo2.UserGetByID(context.Background(), 100)
	suite.NoError(err)
	suite.Equal(100, int(actual.ID))

	// Пытаемся записать в закрытый репозиторий
	suite.NoError(repo2.Close())
	suite.Equal(ErrAOFWrite, repo2.UserCreate(context.Background(), &models.User{ID: 1000}))
	// Проверяем, что пользователь не записан в репозиторий
	_, err = repo2.UserGetByID(context.Background(), 1000)
	suite.Equal(ErrNotFound, err)
}

func (suite *aofRepoSuite) TestAOFRepo_ShortURLCreate() {
	// Создаем репозиторий и записываем в него сокращенную ссылку
	repo1, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(repo1.Close())

	// Открываем репозиторий и проверяем, что сокращенная ссылка записана в него
	repo2, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	actual, err := repo2.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.Equal(suite.testShortURLs[0].ID, actual.ID)
	suite.Equal(suite.testShortURLs[0].OriginalURL, actual.OriginalURL)
	suite.Equal(suite.testShortURLs[0].UserID, actual.UserID)

	// Пытаемся записать в закрытый репозиторий
	suite.NoError(repo2.Close())
	suite.Equal(ErrAOFWrite, repo2.ShortURLCreate(context.Background(), suite.testShortURLs[3]))
	// Проверяем, что сокращенная ссылка не записана в репозиторий
	_, err = repo2.ShortURLGetByID(context.Background(), suite.testShortURLs[3].ID)
	suite.Equal(ErrNotFound, err)
	// Проверяем, что сокращенная ссылка также не доступна в списке ссылок пользователя
	userURLs, err := repo2.ShortURLGetByUserID(context.Background(), suite.testShortURLs[3].UserID)
	suite.NoError(err)
	suite.Equal(0, len(userURLs))
}

func (suite *aofRepoSuite) TestShortURLDelete() {
	// Создаем репозиторий и записываем в него сокращенные ссылки
	repo1, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[1]))
	suite.NoError(repo1.Close())

	// Открываем репозиторий и проверяем, что сокращенные ссылки записаны в него
	repo2, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	actual, err := repo2.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.Equal(suite.testShortURLs[0], actual)
	actual, err = repo2.ShortURLGetByID(context.Background(), suite.testShortURLs[1].ID)
	suite.NoError(err)
	suite.Equal(suite.testShortURLs[1], actual)

	// Помечаем ссылку как удаленную
	suite.NoError(repo2.ShortURLDelete(context.Background(), suite.testShortURLs[0].UserID, suite.testShortURLs[0].ID))
	// Также пытаемся пометить как удаленную ссылку не принадлежащую пользователю
	suite.Equal(ErrNotFound, repo2.ShortURLDelete(context.Background(), 9999, suite.testShortURLs[1].ID))
	suite.NoError(repo2.Close())

	// Проверяем, что первая ссылка помечена как удаленная, а вторая нет
	repo3, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)

	actual, err = repo3.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = repo3.ShortURLGetByID(context.Background(), suite.testShortURLs[1].ID)
	suite.NoError(err)
	suite.Equal(false, actual.Deleted, "should not be deleted")
}

func (suite *aofRepoSuite) TestShortURLDeleteBatch() {
	// Создаем репозиторий и записываем в него сокращенные ссылки
	repo1, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[0]))
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[1]))
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[2]))
	suite.NoError(repo1.ShortURLCreate(context.Background(), suite.testShortURLs[3]))
	suite.NoError(repo1.Close())

	// Открываем репозиторий и проверяем, что сокращенные ссылки записаны в него
	repo2, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	result, err := repo2.ShortURLGetByUserID(context.Background(), suite.testShortURLs[0].UserID)
	suite.NoError(err)
	suite.Equal(3, len(result))
	suite.Equal(suite.testShortURLs[0], &result[0])
	suite.Equal(suite.testShortURLs[1], &result[1])
	suite.Equal(suite.testShortURLs[2], &result[2])
	result, err = repo2.ShortURLGetByUserID(context.Background(), suite.testShortURLs[3].UserID)
	suite.NoError(err)
	suite.Equal(1, len(result))
	suite.Equal(suite.testShortURLs[3], &result[0])

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

	num, err := repo2.ShortURLDeleteBatch(context.Background(), 1, chA, chB)
	suite.NoError(err)
	suite.Equal(3, int(num))

	// Открываем репозиторий и проверяем, что нужные ссылки помечены как удаленные
	repo3, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)

	actual, err := repo3.ShortURLGetByID(context.Background(), suite.testShortURLs[0].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = repo3.ShortURLGetByID(context.Background(), suite.testShortURLs[1].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = repo3.ShortURLGetByID(context.Background(), suite.testShortURLs[2].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(true, actual.Deleted, "should be deleted")

	actual, err = repo3.ShortURLGetByID(context.Background(), suite.testShortURLs[3].ID)
	suite.NoError(err)
	suite.NotNil(actual)
	suite.Equal(false, actual.Deleted, "should not be deleted")
}

func TestAOFRepo(t *testing.T) {
	suite.Run(t, new(aofRepoSuite))
}

// invalidJSONFile - создает файл с некорректными данными (не JSON)
func (suite *aofRepoSuite) invalidJSONFile() {
	f, err := os.OpenFile(suite.filePath, os.O_CREATE|os.O_WRONLY, 0644)
	suite.NoError(err)
	_, err = f.Write([]byte(strings.Repeat("This is not a valid JSON file!\n", 10)))
	suite.NoError(err)
	suite.NoError(f.Close())
}

// invalidJSONStruct - создает файл с некорректными JSON-структурами
func (suite *aofRepoSuite) invalidJSONStruct() {
	f, err := os.OpenFile(suite.filePath, os.O_CREATE|os.O_WRONLY, 0644)
	suite.NoError(err)
	_, err = f.Write([]byte(strings.Repeat(`{"invalid": "key", "wrong": "value"}`+"\n", 10)))
	suite.NoError(err)
	suite.NoError(f.Close())
}

// duplicateUser - создает файл с одинаковыми ID пользователей
func (suite *aofRepoSuite) duplicateUser() {
	userStr := `{"user_create":{"id":100}}` + "\n"
	f, err := os.OpenFile(suite.filePath, os.O_CREATE|os.O_WRONLY, 0644)
	suite.NoError(err)
	_, err = f.Write([]byte(strings.Repeat(userStr, 2)))
	suite.NoError(err)
	suite.NoError(f.Close())
}

// duplicateShortURL - создает файл с одинаковыми ID сокращенных ссылок
func (suite *aofRepoSuite) duplicateShortURL() {
	userStr := `{"user_create":{"id":100}}` + "\n"
	shortURLStr := `{"short_url_create":{"id":"abc123","original_url":"https://www.ya.ru","user_id":100}}` + "\n"
	f, err := os.OpenFile(suite.filePath, os.O_CREATE|os.O_WRONLY, 0644)
	suite.NoError(err)
	_, err = f.Write([]byte(userStr))
	suite.NoError(err)
	_, err = f.Write([]byte(strings.Repeat(shortURLStr, 2)))
	suite.NoError(err)
	suite.NoError(f.Close())
}
