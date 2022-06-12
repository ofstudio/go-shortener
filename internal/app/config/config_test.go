package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewFromEnv(t *testing.T) {
	t.Run("no env", func(t *testing.T) {
		cfg, err := NewFromEnv()
		require.NoError(t, err)
		require.Equal(t, defaultConfig.BaseURL, cfg.BaseURL)
		require.Equal(t, defaultConfig.ServerAddress, cfg.ServerAddress)
	})

	t.Run("with env", func(t *testing.T) {
		_ = os.Setenv("BASE_URL", "https://example.com/")
		_ = os.Setenv("SERVER_ADDRESS", "10.10.0.1:3000")
		cfg, err := NewFromEnv()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		require.Equal(t, "10.10.0.1:3000", cfg.ServerAddress)
	})

	t.Run("BASE_URL without slash", func(t *testing.T) {
		_ = os.Setenv("BASE_URL", "https://example.com")
		cfg, err := NewFromEnv()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", cfg.BaseURL)
		_ = os.Setenv("BASE_URL", "https://example.com/a/b")
		cfg, err = NewFromEnv()
		require.NoError(t, err)
		require.Equal(t, "https://example.com/a/b/", cfg.BaseURL)
	})

	t.Run("invalid BASE_URL", func(t *testing.T) {
		_ = os.Setenv("BASE_URL", "invalid")
		_, err := NewFromEnv()
		require.Error(t, err)
	})

	t.Run("BASE_URL with query", func(t *testing.T) {
		_ = os.Setenv("BASE_URL", "https://example.com/?foo=bar")
		_, err := NewFromEnv()
		require.Error(t, err)
	})
}
