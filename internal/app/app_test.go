package app

import (
	"github.com/ofstudio/go-shortener/internal/storage"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestApp_validateURL(t *testing.T) {
	ulrMaxLen := 20
	publicURL := "https://example.com"
	cfg := NewConfig(ulrMaxLen, publicURL)
	a := NewApp(cfg, storage.NewMemory())

	t.Run("valid URL", func(t *testing.T) {
		got := a.validateURL("https://example.com")
		require.NoError(t, got)
	})

	t.Run("invalid URL", func(t *testing.T) {
		got := a.validateURL("file:///etc/passwd")
		require.Error(t, got)
	})

	t.Run("invalid URL length", func(t *testing.T) {
		got := a.validateURL("https://example.com/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/")
		require.Error(t, got)
	})

	t.Run("invalid URL scheme", func(t *testing.T) {
		got := a.validateURL("ftp://example.com")
		require.Error(t, got)
	})
}

func TestApp_CreateShortURL(t *testing.T) {
	ulrMaxLen := 1024
	publicURL := "https://example.com"
	cfg := NewConfig(ulrMaxLen, publicURL)
	a := NewApp(cfg, storage.NewMemory())

	t.Run("successful url short", func(t *testing.T) {
		got, err := a.CreateShortURL("https://example.com/")
		require.NoError(t, err)
		u, err := url.Parse(got)
		require.NoError(t, err)
		require.Equal(t, publicURL, u.Scheme+"://"+u.Host)
		require.Greater(t, len(u.Path), 1)
	})
}

func TestApp_GetLongURL(t *testing.T) {
	ulrMaxLen := 1024
	publicURL := "https://example.com"
	cfg := NewConfig(ulrMaxLen, publicURL)
	a := NewApp(cfg, storage.NewMemory())

	t.Run("successful get long url", func(t *testing.T) {
		shortURL, err := a.CreateShortURL("https://example.com/")
		require.NoError(t, err)
		u, err := url.Parse(shortURL)
		require.NoError(t, err)
		id := u.Path[1:]
		got, err := a.GetLongURL(id)
		require.NoError(t, err)
		require.Equal(t, "https://example.com/", got)
	})

	t.Run("not found", func(t *testing.T) {
		got, err := a.GetLongURL("not_found")
		require.Equal(t, ErrURLNotFound, err)
		require.Empty(t, got)
	})
}
