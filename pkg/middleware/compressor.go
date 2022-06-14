package middleware

import (
	"net/http"
	"strings"
)

// Compressor - реализация middleware для сжатия ответов с помощью gzip.
// Если ответ имеет допустимый Content-Type и его размер первышает minSize, то он будет сжат.
// В противном случае ответ будет отправлен несжатым
type Compressor struct {
	allowedTypes map[string]struct{} // Список типов, которые могут быть сжаты (если не задано ни одного, то сжимаем все типы)
	minSize      int64               // Минимальный размер данных для сжатия
	level        int                 // Уровень сжатия данных
}

// NewCompressor - создаёт новый компрессор.
// Параметры:
//    minSize - минимальный размер данных для сжатия
//    level - уровень сжатия данных (например, gzip.BestSpeed)
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
		// Если не выполняются условия для сжатия ответа, передаём запрос дальше
		if !c.canCompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Создаём CompressedWriter
		cw := NewCompressedWriter(w, c.minSize, c.level)

		// Необходимо закрыть компрессор после завершения обработки запроса,
		// тк в его буфере могут быть неотправленные данные.
		defer func() {
			if err := cw.Close(); err != nil {
				respondWithError(w, err)
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

// canCompress - проверяет, может ли ответ сжиматься.
// Проверется заголовок Accept-Encoding на наличие gzip
// и Content-Type на список разрешенных для сжатия типов.
func (c *Compressor) canCompress(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") &&
		c.isAllowedType(r.Header.Get("Content-Type"))
}

// isAllowedType - проверяет, может ли сжиматься данный Content-Type.
func (c *Compressor) isAllowedType(contentType string) bool {
	// Если не задано ни одного типа, сжимаем по умолчанию все типы.
	if len(c.allowedTypes) == 0 {
		return true
	}
	_, ok := c.allowedTypes[contentType]
	return ok
}
