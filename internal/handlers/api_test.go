package handlers

import (
	"encoding/json"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/storage"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIHandlers_CreateShortURL(t *testing.T) {
	cfg := &config.Config{URLMaxLen: 20, PublicURL: "https://example.com/"}
	srv := services.NewShortenerService(cfg, storage.NewMemoryStorage())

	t.Run("successful", func(t *testing.T) {
		reqBody := strings.NewReader(`{"url":"https://me.com/"}`)
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, resBody := testRequest(t, ts, http.MethodPost, "/shorten", "application/json", reqBody)
		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, "application/json", res.Header.Get("Content-Type"))
		shortURLRes := createShortURLRes{}
		require.NoError(t, json.Unmarshal([]byte(resBody), &shortURLRes))
		require.NotEmpty(t, shortURLRes.Result)
	})

	t.Run("invalid url", func(t *testing.T) {
		reqBody := strings.NewReader(`{"url":"file:///etc/passwd"}`)
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "application/json", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("too long url", func(t *testing.T) {
		reqBody := strings.NewReader(`{"url":"https://example.com/a/b/c/d/e/f/g/h/i/j/k/l/m"}`)
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "application/json", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("empty body", func(t *testing.T) {
		reqBody := strings.NewReader("")
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "application/json", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("empty JSON", func(t *testing.T) {
		reqBody := strings.NewReader("{}")
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "application/json", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		reqBody := strings.NewReader("<this>is not</valid>json")
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "application/json", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("no content type", func(t *testing.T) {
		reqBody := strings.NewReader(`{"url":"https://me.com/"}`)
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid content type", func(t *testing.T) {
		reqBody := strings.NewReader(`{"url":"https://me.com/"}`)
		r := NewAPIHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/shorten", "application/xml", reqBody)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
