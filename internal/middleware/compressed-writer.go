package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
)

// MTUSize - оптимизированное под размер сетевого пакета значение для CompressedWriter.minSize.
// Сжимать данные небольшого размера, длина которых меньше длины сетевого пакета (MTU) — неэффективно.
// Тк, с одной стороны, эффективность сжатия коротких данных небольшая,
// а с другой — для их передачи в любом случае будет использован 1 сетевой пакет целиком.
const MTUSize int64 = 1460

// compressState - состояние компрессора
type compressState int8

const (
	stateUnset    compressState = iota // Решение о сжатии еще принято: не определен тип и/или размер данных (начальное состояние).
	stateCompress                      // Данные нужно сжимать.
	statePass                          // Данные сжимать не нужно.
)

const NoStatusCode = -1

// CompressedWriter - компрессор для gzip-сжатия данных.
// Данные будут сжиматься при соблюдении следующих условий:
//    1. Размер данных для сжатия больше или равен minSize.
//    2. Тип данных будет включен в список allowedTypes.
// Если хотя бы одно из условий не выполняется, то данные не будут сжаты.
type CompressedWriter struct {
	// Состояние компрессора.
	state compressState

	// Результирующий поток.
	http.ResponseWriter

	// Поток для сжатия данных. При создании объекта устанавливается в nil.
	// Инициализируется только если нужно сжимать данные.
	compWriter io.WriteCloser

	// Минимальный размер данных для сжатия.
	// 0 - сжимать данные в любом случае.
	// MTUSize - оптимизированное значение под размер сетевого пакета.
	minSize int64

	// Буфер размера minSize для накопления данных, перед принятием решения о сжатии.
	buf []byte

	// Кол-во данных в буфере.
	buffered int64

	// Список типов, которые могут быть сжаты (если не задано ни одного, то сжимаем все типы)
	allowedTypes map[string]struct{}

	// Проверен ли тип данных в запросе.
	typeChecked bool

	// Уровень сжатия данных
	level int

	// HTTP-код ответа, полученный от WriteHeader.
	// Если WriteHeader не вызывался, имеет значение NoStatusCode.
	// До момента пока не будет определено, нужно сжимать данные или нет,
	// мы не отправляем код в http.ResponseWriter, а храним его в этом поле.
	// Перед первой отправкой данных в http.ResponseWriter или в compWriter,
	// необходимо отправить код с помощью метода resumeWriteHeader.
	statusCode int
}

// NewCompressedWriter - создает новый поток для сжатия данных.
// Параметры:
//    responseWriter - результирующий записывающий поток.
//    minSize - минимальный размер данных для сжатия.
//    level - уровень сжатия данных.
//    allowedTypes - список типов, которые могут быть сжаты (если не задано ни одного, то сжимаем все типы)
func NewCompressedWriter(responseWriter http.ResponseWriter, minSize int64, level int, allowedTypes map[string]struct{}) *CompressedWriter {
	return &CompressedWriter{
		state:          stateUnset,
		ResponseWriter: responseWriter,
		compWriter:     nil,
		minSize:        minSize,
		buf:            make([]byte, minSize),
		buffered:       0,
		allowedTypes:   allowedTypes,
		typeChecked:    false,
		level:          level,
		statusCode:     NoStatusCode,
	}
}

func (w *CompressedWriter) Write(p []byte) (int, error) {

	// Если не была произведена проверка типа данных, то проверяем его.
	if !w.typeChecked {
		// Если тип данных не разрешен для сжатия, то устанавливаем состояние компрессора не сжимать данные
		if !w.typeCheck(w.ResponseWriter.Header().Get("Content-Type")) {
			w.state = statePass
			w.resumeWriteHeader()
		}
	}

	switch {
	// Если было установлено состояние "Данные сжимать не нужно",
	// то отправляем несжатые данные в результирующий поток ResponseWriter
	case w.state == statePass:
		return w.ResponseWriter.Write(p)

	// Если было установлено состояние "Данные нужно сжимать",
	// то отправляем данные для сжатия в поток compWriter.
	case w.state == stateCompress:
		return w.compWriter.Write(p)

	// Если решение о сжатии еще принято...

	// Если ранее полученные данные в буфере + новые данные
	// превышают минимальный размер, то:
	//   - создаем поток для сжатия данных
	//   - устанавливаем состояние "Данные нужно сжимать"
	//   - устанавливаем заголовки ответа
	//   - отправляем в него ранее полученные данные из буфера + новые данные
	case w.buffered+int64(len(p)) > w.minSize:
		var err error
		w.compWriter, err = gzip.NewWriterLevel(w.ResponseWriter, w.level)
		if err != nil {
			return 0, err
		}
		w.state = stateCompress
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
		w.resumeWriteHeader()
		return w.compWriter.Write(append(w.buf[:w.buffered], p...))

	// Если уже имеющиеся данные в буфере + новые данные
	// не превышают минимальный размер для сжатия,
	// то добавляем новые данные в буфер.
	default:
		copy(w.buf[w.buffered:], p)
		w.buffered += int64(len(p))
		return len(p), nil
	}
}

// WriteHeader - отправляет заголовок HTTP-ответа с соответствующим statusCode
// До момента, пока не будет определено, будут сжаты данные или нет,
// мы не отправляем этот код в поток ResponseWriter.
func (w *CompressedWriter) WriteHeader(statusCode int) {
	switch w.state {
	case stateUnset:
		w.statusCode = statusCode
	default:
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

// Close - закрывает поток для сжатия данных.
// Необходимо вызывать Close после завершения отправки ответа,
// тк в буфере могут быть неотправленные данные.
func (w *CompressedWriter) Close() error {
	switch {
	// Если установлено состояние "Данные нужно сжимать",
	// то закрываем поток для сжатия.
	case w.state == stateCompress && w.compWriter != nil:
		return w.compWriter.Close()

	// Если признак решения о сжатии не установлен и есть данные в буфере,
	// то отправляем их в поток для несжатых данных.
	case w.buffered > 0:
		w.resumeWriteHeader()
		_, err := w.ResponseWriter.Write(w.buf[:w.buffered])
		return err

	// Если признак решения о сжатии не установлен и нет данных в буфере,
	// то ничего не делаем.
	default:
		return nil
	}
}

// typeCheck - проверяет, может ли сжиматься данный Content-Type.
func (w *CompressedWriter) typeCheck(contentType string) bool {
	w.typeChecked = true
	// Если не задано ни одного типа, сжимаем по умолчанию все типы.
	if len(w.allowedTypes) == 0 {
		return true
	}
	_, ok := w.allowedTypes[contentType]
	return ok
}

// resumeWriteHeader - возобновляет отправку HTTP-заголовка,
// в случае если он был установлен раннее с помощью WriteHeader.
// Метод необходимо вызывать перед первой отправкой данных в http.ResponseWriter или compWriter.
func (w *CompressedWriter) resumeWriteHeader() {
	if w.statusCode != NoStatusCode {
		w.ResponseWriter.WriteHeader(w.statusCode)
		// Исключаем повторную отправку.
		w.statusCode = NoStatusCode
	}
}
