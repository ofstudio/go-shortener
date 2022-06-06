package handlers

import (
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/storage"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestShortenerHandlers_GetLongURLSh(t *testing.T) {
	cfg := &config.Config{UrlMaxLen: 1024, PublicURL: "https://example.com/"}
	srv := services.NewShortenerService(cfg, storage.NewMemoryStorage())

	shortURL, err := srv.CreateShortURL("https://me.com/")
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
		{
			name: "another wrong id",
			id:   "../etc/passwd",
			want: want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewShortenerHandlers(srv).Routes()
			ts := httptest.NewServer(r)
			defer ts.Close()

			res, _ := testRequest(t, ts, http.MethodGet, "/"+tt.id, nil)

			// statictest_workaround: res.Body уже закрыта на выходе из testRequest
			defer res.Body.Close()

			require.Equal(t, tt.want.statusCode, res.StatusCode)
			require.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}

func TestShortenerHandlers_CreateShortURL(t *testing.T) {
	cfg := &config.Config{UrlMaxLen: 20, PublicURL: "https://example.com/"}
	srv := services.NewShortenerService(cfg, storage.NewMemoryStorage())

	t.Run("successful", func(t *testing.T) {
		r := NewShortenerHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, shortURL := testRequest(t, ts, http.MethodPost, "/", strings.NewReader("https://me.com/"))

		// statictest_workaround: res.Body уже закрыта на выходе из testRequest
		defer res.Body.Close()

		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, res.Header.Get("Content-Type"), "text/plain; charset=utf-8")

		u, err := url.Parse(shortURL)
		require.NoError(t, err)
		id := u.Path[1:]

		longURL, err := srv.GetLongURL(id)
		require.NoError(t, err)
		require.Equal(t, "https://me.com/", longURL)
	})

	t.Run("invalid url", func(t *testing.T) {
		r := NewShortenerHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/", strings.NewReader("file:///etc/passwd"))

		// statictest_workaround: res.Body уже закрыта на выходе из testRequest
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("too long url", func(t *testing.T) {
		r := NewShortenerHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/", strings.NewReader("https://me.com/a/b/c/d/e/f/g/h/i/j/k/l/m/"))

		// statictest_workaround: res.Body уже закрыта на выходе из testRequest
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("empty body", func(t *testing.T) {
		r := NewShortenerHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/", nil)

		// statictest_workaround: res.Body уже закрыта на выходе из testRequest
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		r := NewShortenerHandlers(srv).Routes()
		ts := httptest.NewServer(r)
		defer ts.Close()
		res, _ := testRequest(t, ts, http.MethodPost, "/non-existing", strings.NewReader("https://me.com/"))

		// statictest_workaround: res.Body уже закрыта на выходе из testRequest
		defer res.Body.Close()

		require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	})
}

func TestShortenerHandlers_notAllowedHTTPMethods(t *testing.T) {
	cfg := &config.Config{UrlMaxLen: 1024, PublicURL: "https://example.com/"}
	srv := services.NewShortenerService(cfg, storage.NewMemoryStorage())

	r := NewShortenerHandlers(srv).Routes()
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, method := range []string{
		http.MethodHead,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	} {
		res, _ := testRequest(t, ts, method, "/", nil)

		// statictest_workaround: res.Body уже закрыта на выходе из testRequest
		_ = res.Body.Close()

		require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	}
}

// testRequest - общая функция для отправки тестовых запросов
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	// HTTP клиент, который не переходит по редиректам
	// https://stackoverflow.com/a/38150816
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	defer res.Body.Close()

	return res, string(respBody)
}
