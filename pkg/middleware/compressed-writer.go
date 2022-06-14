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

// CompressedWriter - реализация потока для gzip-сжатия данных.
// Если совокупный размер полученных данных не превышает минимальный размер minSize,
// то данные сжиматься не будут.
type CompressedWriter struct {
	// Результирующий записывающий поток.
	// Если не принято решение сжимать данные, то записываем данные непосредственно в него.
	http.ResponseWriter

	// Поток для сжатия данных. При создании объекта устанавливается в nil.
	// Инициализируется только в случае принятия решения о сжатии данных.
	compWriter io.WriteCloser

	// Минимальный размер данных для сжатия.
	// 0 - сжимать данные в любом случае.
	// MTUSize - оптимизированное значение под размер сетевого пакета.
	minSize int64

	// Буфер размера minSize для накопления данных, перед принятем решения о сжатии.
	buf []byte

	// Кол-во данных в буфере.
	buffered int64

	// Признак, что принято решение сжимать данные, тк их длина превысила минимальный размер minSize.
	shouldCompress bool

	// Уровень сжатия данных
	level int
}

// NewCompressedWriter - создает новый поток для сжатия данных.
// Параметры:
//    responseWriter - результирующий записывающий поток.
//    minSize - минимальный размер данных для сжатия.
//    level - уровень сжатия данных.
func NewCompressedWriter(responseWriter http.ResponseWriter, minSize int64, level int) *CompressedWriter {
	return &CompressedWriter{
		ResponseWriter: responseWriter,
		compWriter:     nil,
		minSize:        minSize,
		buf:            make([]byte, minSize),
		buffered:       0,
		shouldCompress: false,
		level:          level,
	}
}

func (w *CompressedWriter) Write(p []byte) (int, error) {
	switch {
	// Если ранее уже был установлен признак решения о сжатии,
	// то отправляем данные в поток для сжатия.
	case w.shouldCompress:
		return w.compWriter.Write(p)

	// Если ранее полученные данные в буфере + новые данные
	// превышают минимальный размер, то:
	//   - создаем поток для сжатия данных,
	//   - устанавливаем признак, что принято решение о сжатии,
	//   - устанавливаем заголовки ответа,
	//   - отправляем в него ранее полученные данные из буфера + новые данные.
	case w.buffered+int64(len(p)) > w.minSize:
		var err error
		w.compWriter, err = gzip.NewWriterLevel(w.ResponseWriter, w.level)
		if err != nil {
			return 0, err
		}
		w.shouldCompress = true
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
		return w.compWriter.Write(append(w.buf[:w.buffered], p...))

	// Если ранее полученные данные в буфере + новые данные
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
	// Если признак решения о сжатии установлен,
	// то закрываем поток для сжатия.
	case w.shouldCompress:
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
