package middleware

import (
	"compress/gzip"
	"net/http"
)

// Decompressor - middleware, который распаковывает звпросы с Content-Encoding: gzip.
// Заголовок Content-Length удаляется, тк его значение после распковки не известно.
func Decompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если Content-Encoding не gzip - передаём запрос дальше
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}
		// Распаковываем запрос
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			respondWithError(w, err)
			return
		}
		//goland:noinspection ALL
		defer gzipReader.Close()
		// После распаковки запроса Content-Length будет неопределённым
		r.Header.Del("Content-Length")
		r.Body = gzipReader
		next.ServeHTTP(w, r)
	})
}
