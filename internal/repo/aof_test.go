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
	tmpDir   string
	filePath string
}

func (suite *aofRepoSuite) SetupTest() {
	var err error
	suite.tmpDir, err = os.MkdirTemp("", "aof_test-*")
	suite.NoError(err)
	suite.filePath = suite.tmpDir + "/shortener.aof"
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
	user, err := repo2.UserGetByID(context.Background(), 100)
	suite.NoError(err)
	suite.Equal(100, int(user.ID))

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
	suite.NoError(repo1.ShortURLCreate(context.Background(), &models.ShortURL{ID: "abc123", OriginalURL: "https://www.ya.ru", UserID: 100}))
	suite.NoError(repo1.Close())

	// Открываем репозиторий и проверяем, что сокращенная ссылка записана в него
	repo2, err := NewAOFRepo(suite.filePath)
	suite.NoError(err)
	shortURL, err := repo2.ShortURLGetByID(context.Background(), "abc123")
	suite.NoError(err)
	suite.Equal("abc123", shortURL.ID)
	suite.Equal("https://www.ya.ru", shortURL.OriginalURL)
	suite.Equal(100, int(shortURL.UserID))

	// Пытаемся записать в закрытый репозиторий
	suite.NoError(repo2.Close())
	suite.Equal(ErrAOFWrite, repo2.ShortURLCreate(context.Background(), &models.ShortURL{ID: "xyz123", OriginalURL: "https://www.ya.ru", UserID: 1000}))
	// Проверяем, что сокращенная ссылка не записана в репозиторий
	_, err = repo2.ShortURLGetByID(context.Background(), "xyz123")
	suite.Equal(ErrNotFound, err)
	// Проверяем, что сокращенная ссылка также не доступна в списке ссылок пользователя
	userURLs, err := repo2.ShortURLGetByUserID(context.Background(), 1000)
	suite.NoError(err)
	suite.Equal(0, len(userURLs))
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
