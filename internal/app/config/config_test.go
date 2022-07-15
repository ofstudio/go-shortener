package config

import (
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type configSuite struct {
	suite.Suite
}

func (suite *configSuite) SetupTest() {
	os.Clearenv()
}

func (suite *configSuite) TestNewFromEnv_all() {
	// Устанавливаем все переменные окружения
	suite.setenv(map[string]string{
		"AUTH_TTL":          "1h",
		"AUTH_SECRET":       "secret",
		"BASE_URL":          "https://example.com/",
		"SERVER_ADDRESS":    "localhost:8888",
		"FILE_STORAGE_PATH": "/tmp/shortener.aof",
	})

	cfg, err := NewFromEnv()
	suite.NoError(err)

	// Проверяем, что все переменные окружения прочитаны
	suite.Equal(time.Hour*1, cfg.AuthTTL)
	suite.Equal("secret", cfg.AuthSecret)
	suite.Equal(mustParseRequestURI("https://example.com/"), cfg.BaseURL)
	suite.Equal("localhost:8888", cfg.ServerAddress)
	suite.Equal("/tmp/shortener.aof", cfg.FileStoragePath)
}

func (suite *configSuite) TestNewFromEnv_partial() {
	// Устанавливаем только часть переменных окружения
	suite.setenv(map[string]string{
		"AUTH_TTL":          "1h",
		"BASE_URL":          "https://example.com/",
		"FILE_STORAGE_PATH": "/tmp/shortener.aof",
	})

	cfg, err := NewFromEnv()
	suite.NoError(err)

	// Проверяем, что прочитаны заданные переменные окружения
	suite.Equal(time.Hour*1, cfg.AuthTTL)
	suite.Equal("https://example.com/", cfg.BaseURL.String())
	suite.Equal("/tmp/shortener.aof", cfg.FileStoragePath)

	// Проверяем, что остальные параметры установлены в значения по умолчанию
	suite.Equal(DefaultConfig.AuthSecret, cfg.AuthSecret)
	suite.Equal(DefaultConfig.ServerAddress, cfg.ServerAddress)
}

func (suite *configSuite) TestNewFromEnvAndCLI_all() {
	// Устанавливаем все доступные флаги
	args := []string{
		"-a", "127.0.0.0:8888",
		"-b", "https://example.com/",
		"-f", "/tmp/shortener.aof",
		"-t", "1h",
	}

	cfg, err := newFromEnvAndCLI(args)
	suite.NoError(err)

	// Проверяем, что прочитаны все заданные флаги
	suite.Equal("127.0.0.0:8888", cfg.ServerAddress)
	suite.Equal("https://example.com/", cfg.BaseURL.String())
	suite.Equal("/tmp/shortener.aof", cfg.FileStoragePath)
	suite.Equal(time.Hour*1, cfg.AuthTTL)

	// Проверяем, что остальные параметры установлены в значения по умолчанию
	suite.Equal(DefaultConfig.AuthSecret, cfg.AuthSecret)
}

func (suite *configSuite) TestNewFromEnvAndCLI_partial() {
	// Устанавливаем часть доступных флагов
	args := []string{
		"-a", "0.0.0.0:3000",
		"-d", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
	}
	// Устанавливаем часть переменных окружения
	suite.setenv(map[string]string{
		"FILE_STORAGE_PATH": "/tmp/shortener.aof",
		"BASE_URL":          "https://example.com/",
	})

	cfg, err := newFromEnvAndCLI(args)
	suite.NoError(err)

	// Проверяем, что прочитаны заданные флаги
	suite.Equal("0.0.0.0:3000", cfg.ServerAddress)
	suite.Equal("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", cfg.DatabaseDSN)

	// Проверяем, что прочитаны заданные переменные окружения
	suite.Equal("/tmp/shortener.aof", cfg.FileStoragePath)
	suite.Equal("https://example.com/", cfg.BaseURL.String())

	// Проверяем, что остальные параметры установлены в значения по умолчанию
	suite.Equal(DefaultConfig.AuthSecret, cfg.AuthSecret)
	suite.Equal(DefaultConfig.AuthTTL, cfg.AuthTTL)
}

func (suite *configSuite) TestValidateBaseURL() {
	// Проверяем на невалидный URL
	_, err := newFromEnvAndCLI([]string{"-a", "not-a-valid-url"})
	suite.Error(err)
	suite.setenv(map[string]string{"BASE_URL": "ftps://example.com/"})
	_, err = newFromEnvAndCLI([]string{})
	suite.Error(err)

	// Проверяем, что к URL добавляется слеш в конце
	suite.setenv(map[string]string{"BASE_URL": "https://example.com"})
	cfg, err := newFromEnvAndCLI([]string{})
	suite.NoError(err)
	suite.Equal("https://example.com/", cfg.BaseURL.String())
	cfg, err = newFromEnvAndCLI([]string{"-b", "https://example.com/subpath"})
	suite.NoError(err)
	suite.Equal("https://example.com/subpath/", cfg.BaseURL.String())

	// Проверяем что не допускаются URL с параметрами или фрагментами
	suite.setenv(map[string]string{"BASE_URL": "https://example.com/subpath?param=1"})
	_, err = newFromEnvAndCLI([]string{})
	suite.Error(err)
	_, err = newFromEnvAndCLI([]string{"-b", "https://example.com/subpath#fragment"})
	suite.Error(err)
}

func (suite *configSuite) TestValidateServerAddress() {
	// Проверяем на невалидный адрес сервера
	_, err := newFromEnvAndCLI([]string{"-a", "not-a-valid-tcp-address"})
	suite.Error(err)
	suite.setenv(map[string]string{"SERVER_ADDRESS": "0.0.0.0:100000"})
	_, err = newFromEnvAndCLI([]string{})
	suite.Error(err)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(configSuite))
}

// setenv - устанавливает переменные окружения
func (suite *configSuite) setenv(vars map[string]string) {
	os.Clearenv()
	for k, v := range vars {
		suite.NoError(os.Setenv(k, v))
	}
}
