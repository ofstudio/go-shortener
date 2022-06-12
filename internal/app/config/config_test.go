package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewFromEnv(t *testing.T) {
	t.Run("no env", func(t *testing.T) {
		os.Clearenv()
		cfg, err := validate(fromEnv())
		require.NoError(t, err)
		require.Equal(t, DefaultConfig.BaseURL, cfg.BaseURL)
		require.Equal(t, DefaultConfig.ServerAddress, cfg.ServerAddress)
	})

	t.Run("with env", func(t *testing.T) {
		os.Clearenv()
		_ = os.Setenv("BASE_URL", "https://example.com/")
		_ = os.Setenv("SERVER_ADDRESS", "10.10.0.1:3000")
		cfg, err := validate(fromEnv())
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		require.Equal(t, "10.10.0.1:3000", cfg.ServerAddress)
	})

	t.Run("BASE_URL without slash", func(t *testing.T) {
		os.Clearenv()
		_ = os.Setenv("BASE_URL", "https://example.com")
		cfg, err := NewFromEnv()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		_ = os.Setenv("BASE_URL", "https://example.com/a/b")
		cfg, err = validate(fromEnv())
		require.NoError(t, err)
		require.Equal(t, "https://example.com/a/b/", cfg.BaseURL)
	})

	t.Run("invalid BASE_URL", func(t *testing.T) {
		os.Clearenv()
		_ = os.Setenv("BASE_URL", "invalid")
		_, err := validate(fromEnv())
		require.Error(t, err)
	})

	t.Run("BASE_URL with query", func(t *testing.T) {
		os.Clearenv()
		_ = os.Setenv("BASE_URL", "https://example.com/?foo=bar")
		_, err := validate(fromEnv())
		require.Error(t, err)
	})

	// Неверный адрес сервера
	t.Run("invalid ServerAddress", func(t *testing.T) {
		os.Clearenv()
		_ = os.Setenv("SERVER_ADDRESS", "invalid")
		_, err := validate(fromEnv())
		require.Error(t, err)
	})
}

func TestNewFromEnvAndCLI(t *testing.T) {
	// Все параметры заданы через коммандную строку
	t.Run("with CLI", func(t *testing.T) {
		os.Clearenv()
		args := []string{"-a", "127.0.0.0:8888", "-b", "https://example.com/", "-f", "/tmp/shortener.aof"}
		cfg, err := validate(fromEnvAndCLI(args))
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		require.Equal(t, "127.0.0.0:8888", cfg.ServerAddress)
		require.Equal(t, "/tmp/shortener.aof", cfg.FileStoragePath)
	})

	// Через командную строку задан только BaseURL, остальыне параметры используют значения по умолчанию
	t.Run("with CLI only BaseURL", func(t *testing.T) {
		os.Clearenv()
		args := []string{"-b", "https://example.com/"}
		cfg, err := validate(fromEnvAndCLI(args))
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		require.Equal(t, DefaultConfig.ServerAddress, cfg.ServerAddress)
		require.Equal(t, DefaultConfig.FileStoragePath, cfg.FileStoragePath)
	})

	// Через окружение задан BaseUrl и FileStoragePath.
	// Через командную строку задан FileStoragePath.
	// Остальыне параметры используют значения по умолчанию
	t.Run("with env and CLI", func(t *testing.T) {
		os.Clearenv()
		_ = os.Setenv("BASE_URL", "https://example.com/")
		_ = os.Setenv("FILE_STORAGE_PATH", "/tmp/env.aof")
		args := []string{"-f", "/tmp/cli.aof"}
		cfg, err := validate(fromEnvAndCLI(args))
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		require.Equal(t, DefaultConfig.ServerAddress, cfg.ServerAddress)
		require.Equal(t, "/tmp/cli.aof", cfg.FileStoragePath)
	})

	// Проверка на корректный BaseURL
	t.Run("with invalid BaseURL", func(t *testing.T) {
		os.Clearenv()
		args := []string{"-b", "invalid_url"}
		_, err := validate(fromEnvAndCLI(args))
		require.Error(t, err)
	})

	// Проверка на добавление в конец BaseURL слеша
	t.Run("add slash in BaseURL", func(t *testing.T) {
		os.Clearenv()
		args := []string{"-b", "https://example.com/a/b"}
		cfg, err := validate(fromEnvAndCLI(args))
		require.NoError(t, err)
		require.Equal(t, "https://example.com/a/b/", cfg.BaseURL)
	})

	// Неверный адрес сервера
	t.Run("invalid ServerAddress", func(t *testing.T) {
		os.Clearenv()
		args := []string{"-a", "invalid"}
		_, err := validate(fromEnvAndCLI(args))
		require.Error(t, err)
	})

}
