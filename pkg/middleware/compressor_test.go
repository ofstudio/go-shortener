package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompressor(t *testing.T) {
	// Подходит тип, размер, заголовок запроса
	// Заголовок ответа Content-Encoding установлен
	// Декомпрессированный body ответа совпадает c запросом
	t.Run("compressed response", func(t *testing.T) {
		ts := testCompressorServer(t, 0, "application/json")
		defer ts.Close()
		res, resBody := testCompressorRequest(t, ts, "application/json", []byte(`{"a":1}`), true)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, "application/json", res.Header.Get("Content-Type"))
		require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
		require.Equal(t, "Accept-Encoding", res.Header.Get("Vary"))
		require.Equal(t, fmt.Sprintf("%d", len(resBody)), res.Header.Get("Content-Length"))
		require.Equal(t, `{"a":1}`, string(testGzipDecompress(t, resBody)))
	})

	// Сервер сжимает любые типы, подходит размер и заголовок запроса
	// Заголовок ответа Content-Encoding установлен
	// Декомпрессированный body ответа совпадает c запросом
	t.Run("compressed response - server accept any content type", func(t *testing.T) {
		ts := testCompressorServer(t, 0)
		defer ts.Close()
		res, resBody := testCompressorRequest(t, ts, "application/json", []byte(`{"a":1}`), true)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, "application/json", res.Header.Get("Content-Type"))
		require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
		require.Equal(t, "Accept-Encoding", res.Header.Get("Vary"))
		require.Equal(t, fmt.Sprintf("%d", len(resBody)), res.Header.Get("Content-Length"))
		require.Equal(t, `{"a":1}`, string(testGzipDecompress(t, resBody)))
	})

	// Подходит тип, размер, не подходит заголовок запроса
	// Заголовок ответа Content-Encoding не установлен
	// Body ответа не компрессирован и совпадает с запросом
	t.Run("uncompressed response - Accept-Encoding missing", func(t *testing.T) {
		ts := testCompressorServer(t, 0, "text/plain")
		defer ts.Close()
		res, resBody := testCompressorRequest(t, ts, "text/plain", []byte("test"), false)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		require.Equal(t, "", res.Header.Get("Content-Encoding"))
		require.Equal(t, "test", string(resBody))
	})

	// Подходит размер и заголовок запроса, не подходит тип
	// Заголовок ответа Content-Encoding не установлен
	// Body ответа не компрессирован и совпадает с запросом
	t.Run("uncompressed response - Content-Type mismatch", func(t *testing.T) {
		ts := testCompressorServer(t, 0, "text/plain")
		defer ts.Close()
		res, resBody := testCompressorRequest(t, ts, "text/html", []byte("test"), true)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, "text/html", res.Header.Get("Content-Type"))
		require.Equal(t, "", res.Header.Get("Content-Encoding"))
		require.Equal(t, "test", string(resBody))
	})

	// Подходит тип и заголовок запроса, не подходит размер
	// Заголовок ответа Content-Encoding не установлен
	// Body ответа не компрессирован и совпадает с запросом
	t.Run("uncompressed response - body is too short", func(t *testing.T) {
		ts := testCompressorServer(t, MTUSize, "text/plain")
		defer ts.Close()
		res, resBody := testCompressorRequest(t, ts, "text/css", []byte("test"), true)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		require.Equal(t, "text/css", res.Header.Get("Content-Type"))
		require.Equal(t, "", res.Header.Get("Content-Encoding"))
		require.Equal(t, "test", string(resBody))
	})
}

// testCompressorRequest - тестовый запрос
// Параметры:
//   ts - тестовый сервер
//   contentType - тип контента
//   reqBody - тело запроса
//   acceptCompress - принимать ли сжатые ответы
// Возвращает:
//  - ответ от сервера
//	- body ответа, не-декомпрессированное
func testCompressorRequest(t *testing.T, ts *httptest.Server, contentType string, reqBody []byte, acceptCompress bool) (*http.Response, []byte) {
	req, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)
	if acceptCompress {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		require.NoError(t, err)
	}(res.Body)
	resBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	return res, resBody
}

// testCompressorServer - тестовый echo-сервер
// Параметры:
// 	- минимальный размер для сжатия
//  - допустимые типы для сжатия
func testCompressorServer(t *testing.T, minSize int64, types ...string) *httptest.Server {
	compressor := NewCompressor(minSize, gzip.BestSpeed)
	for _, ct := range types {
		compressor.AddType(ct)
	}

	r := chi.NewRouter()
	r.Use(compressor.Handler)

	// Echo-хендлер
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			require.NoError(t, err)
		}(r.Body)
		_, err = w.Write(body)
		require.NoError(t, err)
	})

	return httptest.NewServer(r)
}
