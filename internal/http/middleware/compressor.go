package middleware

import (
	"net/http"
	"strings"
)

// Compressor - реализация middleware для сжатия ответов с помощью gzip.
// Данные будут сжиматься при соблюдении следующих условий:
//  1. Заголовок запроса Content-Encoding содержит gzip
//  2. Размер данных для сжатия больше или равен minSize.
//  3. Тип данных будет включен в список allowedTypes.
//
// Если хотя бы одно из условий не выполняется, то данные не будут сжаты.
type Compressor struct {
	allowedTypes map[string]struct{} // Список типов, которые могут быть сжаты (если не задано ни одного, то сжимаем все типы)
	minSize      int64               // Минимальный размер данных для сжатия
	level        int                 // Уровень сжатия данных
}

// NewCompressor - создаёт новый компрессор.
// Параметры:
//
//	minSize - минимальный размер данных для сжатия
//	level - уровень сжатия данных (например, gzip.BestSpeed)
func NewCompressor(minSize int64, level int) *Compressor {
	return &Compressor{
		allowedTypes: make(map[string]struct{}),
		minSize:      minSize,
		level:        level,
	}
}

// Handler - возвращает middleware для сжатия ответов.
func (c *Compressor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если в Accept-Encoding не указан gzip, то не сжимаем данные
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаём CompressedWriter
		cw := NewCompressedWriter(w, c.minSize, c.level, c.allowedTypes)
		// Необходимо закрыть компрессор после завершения обработки запроса,
		// тк в его буфере могут быть неотправленные данные.
		defer func() {
			if err := cw.Close(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		// Передаём обработчику страницы CompressedWriter для вывода данных.
		next.ServeHTTP(cw, r)
	})
}

// AddType - добавляет тип в список, которые могут быть сжаты.
func (c *Compressor) AddType(contentType string) *Compressor {
	c.allowedTypes[contentType] = struct{}{}
	return c
}
