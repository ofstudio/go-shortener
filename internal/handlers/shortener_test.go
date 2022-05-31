package handlers

import (
	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/storage"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestShortener_get(t *testing.T) {
	ulrMaxLen := 20
	publicURL := "https://example.com"
	cfg := app.NewConfig(ulrMaxLen, publicURL)
	a := app.NewApp(cfg, storage.NewMemory())
	shortURL, err := a.CreateShortURL("https://me.com/")
	require.NoError(t, err)
	u, err := url.Parse(shortURL)
	require.NoError(t, err)
	id := u.Path[1:]

	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name string
		id   string
		want want
	}{
		{
			name: "successful",
			id:   id,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://me.com/",
			},
		},
		{
			name: "not found",
			id:   "wrong-id",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.id)
			req := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
			w := httptest.NewRecorder()
			h := NewShortener(a)
			h.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			require.Equal(t, tt.want.statusCode, res.StatusCode)
			require.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}

func TestShortener_post(t *testing.T) {
	ulrMaxLen := 20
	publicURL := "https://example.com"
	cfg := app.NewConfig(ulrMaxLen, publicURL)
	a := app.NewApp(cfg, storage.NewMemory())

	t.Run("successful", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Body = ioutil.NopCloser(strings.NewReader("https://me.com/"))
		w := httptest.NewRecorder()
		h := NewShortener(a)
		h.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, res.Header.Get("Content-Type"), "text/plain; charset=utf-8")
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		shortUrl := string(body)

		u, err := url.Parse(shortUrl)
		require.NoError(t, err)
		id := u.Path[1:]

		longUrl, err := a.GetLongURL(id)
		require.NoError(t, err)
		require.Equal(t, "https://me.com/", longUrl)
	})

	t.Run("invalid url", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Body = ioutil.NopCloser(strings.NewReader("invalid-url"))
		w := httptest.NewRecorder()
		h := NewShortener(a)
		h.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("too long url", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Body = ioutil.NopCloser(strings.NewReader("https://me.com/a/b/c/d/e/f/g/h/i/j/k/l/m/"))
		w := httptest.NewRecorder()
		h := NewShortener(a)
		h.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		h := NewShortener(a)
		h.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/invalid-endpoint", nil)
		w := httptest.NewRecorder()
		h := NewShortener(a)
		h.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestShortener_ServeHTTP(t *testing.T) {
	ulrMaxLen := 20
	publicURL := "https://example.com"
	cfg := app.NewConfig(ulrMaxLen, publicURL)
	a := app.NewApp(cfg, storage.NewMemory())
	h := NewShortener(a)

	t.Run("not allowed method", func(t *testing.T) {
		notAllowed := []string{
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}

		for _, method := range notAllowed {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			res := w.Result()
			_ = res.Body.Close()
			require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
		}
	})
}
