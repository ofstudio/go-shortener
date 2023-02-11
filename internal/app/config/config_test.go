package config

import (
	"crypto/tls"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
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

	actualCfg, err := FromEnv(suite.defaultCfg())
	suite.NoError(err)

	// Проверяем, что все переменные окружения прочитаны
	suite.Equal(time.Hour*1, actualCfg.AuthTTL)
	suite.Equal("secret", actualCfg.AuthSecret)
	suite.Equal("https://example.com/", actualCfg.BaseURL.String())
	suite.Equal("localhost:8888", actualCfg.ServerAddress)
	suite.Equal("/tmp/shortener.aof", actualCfg.FileStoragePath)
}

func (suite *configSuite) TestNewFromEnv_partial() {
	// Устанавливаем только часть переменных окружения
	suite.setenv(map[string]string{
		"AUTH_TTL":          "1h",
		"BASE_URL":          "https://example.com/",
		"FILE_STORAGE_PATH": "/tmp/shortener.aof",
	})

	defaultCfg := suite.defaultCfg()
	actualCfg, err := FromEnv(defaultCfg)
	suite.NoError(err)

	// Проверяем, что прочитаны заданные переменные окружения
	suite.Equal(time.Hour*1, actualCfg.AuthTTL)
	suite.Equal("https://example.com/", actualCfg.BaseURL.String())
	suite.Equal("/tmp/shortener.aof", actualCfg.FileStoragePath)

	// Проверяем, что остальные параметры установлены в значения по умолчанию
	suite.Equal(defaultCfg.AuthSecret, actualCfg.AuthSecret)
	suite.Equal(defaultCfg.ServerAddress, actualCfg.ServerAddress)
}

func (suite *configSuite) TestNewFromEnvAndCLI_all() {
	// Устанавливаем все доступные флаги
	args := []string{
		"-a", "127.0.0.0:8888",
		"-b", "https://example.com/",
		"-f", "/tmp/shortener.aof",
		"-t", "1h",
	}

	defaultCfg := suite.defaultCfg()
	actualCfg, err := FromCLI(args...)(defaultCfg)
	suite.NoError(err)

	// Проверяем, что прочитаны все заданные флаги
	suite.Equal("127.0.0.0:8888", actualCfg.ServerAddress)
	suite.Equal("https://example.com/", actualCfg.BaseURL.String())
	suite.Equal("/tmp/shortener.aof", actualCfg.FileStoragePath)
	suite.Equal(time.Hour*1, actualCfg.AuthTTL)

	// Проверяем, что остальные параметры установлены в значения по умолчанию
	suite.Equal(defaultCfg.AuthSecret, actualCfg.AuthSecret)
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

	defaultCfg := suite.defaultCfg()
	actualCfg, err := FromEnv(defaultCfg)
	suite.NoError(err)
	actualCfg, err = FromCLI(args...)(actualCfg)
	suite.NoError(err)

	// Проверяем, что прочитаны заданные флаги
	suite.Equal("0.0.0.0:3000", actualCfg.ServerAddress)
	suite.Equal("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", actualCfg.DatabaseDSN)

	// Проверяем, что прочитаны заданные переменные окружения
	suite.Equal("/tmp/shortener.aof", actualCfg.FileStoragePath)
	suite.Equal("https://example.com/", actualCfg.BaseURL.String())

	// Проверяем, что остальные параметры установлены в значения по умолчанию
	suite.Equal(defaultCfg.AuthSecret, actualCfg.AuthSecret)
	suite.Equal(defaultCfg.AuthTTL, actualCfg.AuthTTL)
}

func (suite *configSuite) TestValidateBaseURL() {

	// Проверяем на невалидный URL
	_, err := FromCLI("-a", "not-a-valid-url")(suite.defaultCfg())
	suite.Error(err)
	suite.setenv(map[string]string{"BASE_URL": "ftps://example.com/"})
	_, err = FromEnv(suite.defaultCfg())
	suite.Error(err)

	// Проверяем, что к URL добавляется слеш в конце
	suite.setenv(map[string]string{"BASE_URL": "https://example.com"})
	actualCfg, err := FromEnv(suite.defaultCfg())
	suite.NoError(err)
	suite.Equal("https://example.com/", actualCfg.BaseURL.String())
	actualCfg, err = FromCLI("-b", "https://example.com/subpath")(suite.defaultCfg())
	suite.NoError(err)
	suite.Equal("https://example.com/subpath/", actualCfg.BaseURL.String())

	// Проверяем что не допускаются URL с параметрами или фрагментами
	suite.setenv(map[string]string{"BASE_URL": "https://example.com/subpath?param=1"})
	_, err = FromEnv(suite.defaultCfg())
	suite.Error(err)
	_, err = FromCLI("-b", "https://example.com/subpath#fragment")(suite.defaultCfg())
	suite.Error(err)
}

func (suite *configSuite) TestValidateServerAddress() {
	// Проверяем на невалидный адрес сервера
	_, err := FromCLI("-a", "not-a-valid-tcp-address")(suite.defaultCfg())
	suite.Error(err)
	suite.setenv(map[string]string{"SERVER_ADDRESS": "0.0.0.0:100000"})
	_, err = FromEnv(suite.defaultCfg())
	suite.Error(err)
}

func (suite *configSuite) TestFromJSONFile() {
	suite.Run("valid from cli", func() {
		cfg, err := FromJSONFile("-c", "testdata/cfg-valid.json")(suite.defaultCfg())
		suite.Require().NoError(err)
		suite.Equal("localhost:8080", cfg.ServerAddress)
		suite.Equal("http://localhost/", cfg.BaseURL.String())
		suite.Equal("/path/to/file.db", cfg.FileStoragePath)
		suite.Equal("", cfg.DatabaseDSN)
		suite.Equal(true, cfg.EnableHTTPS)
	})
	suite.Run("valid from env", func() {
		os.Setenv("CONFIG", "testdata/cfg-valid.json")
		cfg, err := FromJSONFile()(suite.defaultCfg())
		suite.Require().NoError(err)
		suite.Equal("localhost:8080", cfg.ServerAddress)
		suite.Equal("http://localhost/", cfg.BaseURL.String())
		suite.Equal("/path/to/file.db", cfg.FileStoragePath)
		suite.Equal("", cfg.DatabaseDSN)
		suite.Equal(true, cfg.EnableHTTPS)
	})
	suite.Run("mistype", func() {
		os.Setenv("CONFIG", "testdata/cfg-mistype.json")
		_, err := FromJSONFile()(suite.defaultCfg())
		suite.Error(err)
	})
	suite.Run("unknown", func() {
		os.Setenv("CONFIG", "testdata/cfg-unknown.json")
		_, err := FromJSONFile()(suite.defaultCfg())
		suite.Error(err)
	})
}

func (suite *configSuite) TestTLS_validate() {
	t := &TLS{
		Hosts: []string{"example.com"},
		Curve: tls.CurveP256,
		TTL:   time.Hour,
	}
	suite.NoError(t.validate())

	t.Hosts = []string{"example.com", "example.org", "127.0.0.1"}
	suite.NoError(t.validate())

	t.Hosts = []string{}
	suite.Error(t.validate())

	t.Hosts = nil
	suite.Error(t.validate())

	t.Hosts = []string{"example.com"}
	t.TTL = 0
	suite.Error(t.validate())

	t.TTL = -1
	suite.Error(t.validate())

	t = &TLS{}
	suite.Error(t.validate())
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

func (suite *configSuite) defaultCfg() *Config {
	cfg, err := Default(nil)
	suite.NoError(err)
	return cfg
}
