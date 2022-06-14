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
	stateCompress                      // Данные необходимо сжимать.
	statePass                          // Данные сжимать не нужно.
)

// CompressedWriter - компрессор для gzip-сжатия данных.
// Если совокупный размер полученных данных не превышает минимальный размер minSize,
// то данные сжиматься не будут.
type CompressedWriter struct {
	// Состояние компрессора.
	state compressState

	// Результирующий поток.
	http.ResponseWriter

	// Поток для сжатия данных. При создании объекта устанавливается в nil.
	// Инициализируется только если необходимо сжимать данные.
	compWriter io.WriteCloser

	// Минимальный размер данных для сжатия.
	// 0 - сжимать данные в любом случае.
	// MTUSize - оптимизированное значение под размер сетевого пакета.
	minSize int64

	// Буфер размера minSize для накопления данных, перед принятем решения о сжатии.
	buf []byte

	// Кол-во данных в буфере.
	buffered int64

	// Список типов, которые могут быть сжаты (если не задано ни одного, то сжимаем все типы)
	allowedTypes map[string]struct{}

	// Проверен ли тип данных в запросе.
	typeChecked bool

	// Уровень сжатия данных
	level int
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
	}
}

func (w *CompressedWriter) Write(p []byte) (int, error) {

	// Если не была произведена проверка типа данных, то проверяем его.
	if !w.typeChecked {
		w.typeChecked = true
		// Если тип данных не разрешен для сжатия, то устанавливаем состояние компрессора не сжимать данные
		if !w.typeCheck(w.ResponseWriter.Header().Get("Content-Type")) {
			w.state = statePass
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
	//   - создаем поток для сжатия данных,
	//   - устанавливаем состояние "Данные нужно сжимать",
	//   - устанавливаем заголовки ответа,
	//   - отправляем в него ранее полученные данные из буфера + новые данные.
	case w.buffered+int64(len(p)) > w.minSize:
		var err error
		w.compWriter, err = gzip.NewWriterLevel(w.ResponseWriter, w.level)
		if err != nil {
			return 0, err
		}
		w.state = stateCompress
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
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
	// Если не задано ни одного типа, сжимаем по умолчанию все типы.
	if len(w.allowedTypes) == 0 {
		return true
	}
	_, ok := w.allowedTypes[contentType]
	return ok
}
