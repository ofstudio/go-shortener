package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecompressor(t *testing.T) {
	// Сервер принимает сжатые запросы.
	// Отпарвляем сжатый запрос.
	// Ответ совпадает с запросом.
	t.Run("accept compressed, sent compressed", func(t *testing.T) {
		ts := testDecompressorServer(t, true)
		defer ts.Close()
		res, resBody := testDecompressorRequest(t, ts, []byte("test"), true)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, "test", string(resBody))
	})

	// Сервер принимает сжатые запросы.
	// Отпарвляем несжатый запрос.
	// Ответ совпадает с запросом.
	t.Run("accept compressed, sent uncompressed", func(t *testing.T) {
		ts := testDecompressorServer(t, true)
		defer ts.Close()
		res, resBody := testDecompressorRequest(t, ts, []byte("test"), false)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, "test", string(resBody))
	})

	// Сервер не принимает сжатые запросы.
	// Отпарвляем сжатый запрос.
	// Ответ не совпадает с запросом.
	// Декомпрессируем ответ и он совпадает с запросом.
	t.Run("don't accept compressed, sent compressed", func(t *testing.T) {
		ts := testDecompressorServer(t, false)
		defer ts.Close()
		res, resBody := testDecompressorRequest(t, ts, []byte("test"), true)
		// statictest_workaround
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.NotEqual(t, "test", string(resBody))
		require.Equal(t, "test", string(testGzipDecompress(t, resBody)))
	})

}

// testDecompressorRequest - тестовый запрос
// Параметры:
// 	  ts - тестовый сервые
//    body - тело запроса
//    compress - сжимать ли запрос или нет
// Возвращает:
//    - ответ от сервера
//	  - body ответа
func testDecompressorRequest(t *testing.T, ts *httptest.Server, reqBody []byte, compress bool) (*http.Response, []byte) {
	if compress {
		reqBody = testGzipCompress(t, reqBody)
	}
	req, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "text/plain")
	require.NoError(t, err)

	if compress {
		req.Header.Set("Content-Encoding", "gzip")
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

// testDecompressorServer -тестовый echo-сервер
// Параметры:
//    acceptCompressed - принимать или нет сжатые запросы
func testDecompressorServer(t *testing.T, acceptCompressed bool) *httptest.Server {
	r := chi.NewRouter()
	if acceptCompressed {
		r.Use(Decompressor)
	}

	// Echo-хендлер
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
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

// testGzipDecompress - декомпрессирует данные
// Параметры:
//    data - данные
// Возвращает:
//    - декомпрессированные данные
func testGzipDecompress(t *testing.T, data []byte) []byte {
	reader := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(reader)
	require.NoError(t, err)
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		require.NoError(t, err)
	}(gzipReader)
	out, err := ioutil.ReadAll(gzipReader)
	require.NoError(t, err)
	return out
}

// testGzipDecompress - сжимает данные
// Параметры:
//    data - данные
// Возвращает:
//    - сжатые данные
func testGzipCompress(t *testing.T, data []byte) []byte {
	buffer := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(buffer)
	_, err := gzipWriter.Write(data)
	require.NoError(t, err)
	err = gzipWriter.Close()
	require.NoError(t, err)
	return buffer.Bytes()
}
