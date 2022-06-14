package services

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/pkg/storage"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestShortenerService_validateURL(t *testing.T) {
	cfg := &config.Config{URLMaxLen: 20, BaseURL: "https://example.com/"}
	srv := NewShortenerService(cfg, storage.NewMemoryStorage())

	t.Run("valid URL", func(t *testing.T) {
		got := srv.validateURL("https://domain.com")
		require.NoError(t, got)
	})

	t.Run("invalid URL", func(t *testing.T) {
		got := srv.validateURL("file:///etc/passwd")
		require.Error(t, got)
	})

	t.Run("invalid URL length", func(t *testing.T) {
		got := srv.validateURL("https://domain.com/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/")
		require.Error(t, got)
	})

	t.Run("invalid URL scheme", func(t *testing.T) {
		got := srv.validateURL("ftp://domain.com")
		require.Error(t, got)
	})
}

func TestShortenerService_CreateShortURL(t *testing.T) {
	cfg := &config.Config{URLMaxLen: 1024, BaseURL: "https://example.com/"}
	srv := NewShortenerService(cfg, storage.NewMemoryStorage())

	t.Run("successful url short", func(t *testing.T) {
		got, err := srv.CreateShortURL("https://example.com/")
		require.NoError(t, err)
		u, err := url.Parse(got)
		require.NoError(t, err)
		require.Equal(t, cfg.BaseURL, u.Scheme+"://"+u.Host+"/")
		require.Greater(t, len(u.Path), 1)
	})
}

func TestShortenerService_GetLongURL(t *testing.T) {
	cfg := &config.Config{URLMaxLen: 1024, BaseURL: "https://example.com/"}
	srv := NewShortenerService(cfg, storage.NewMemoryStorage())

	t.Run("successful get long url", func(t *testing.T) {
		shortURL, err := srv.CreateShortURL("https://domain.com/")
		require.NoError(t, err)
		u, err := url.Parse(shortURL)
		require.NoError(t, err)
		id := u.Path[1:]
		got, err := srv.GetLongURL(id)
		require.NoError(t, err)
		require.Equal(t, "https://domain.com/", got)
	})

	t.Run("not found", func(t *testing.T) {
		got, err := srv.GetLongURL("not_found")
		require.Equal(t, ErrShortURLNotFound, err)
		require.Empty(t, got)
	})
}
