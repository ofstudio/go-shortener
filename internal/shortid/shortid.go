package shortid

import (
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// New - генерирует случайный идентификатор и возвращает его
// в виде URL-friendly (base64) строкового представления.
//
// Идентификатор имеет длину 8 байт и состоит из 2 частей:
//   - 4 байта: текущее Unix-время в секундах
//   - 4 байта: случайное число
func New() string {
	// Текущее Unix-время в секундах
	// Предсавление Unix-секунд в uint32 обеспечивает уникальные значения
	// в течение 136 лет: до 17 фев 2106
	// https://en.wikipedia.org/wiki/Time_formatting_and_storage_bugs#Year_2106
	timePart := uint32(time.Now().Unix())

	// Случайное число
	randPart := rand.Uint32()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, timePart)
	binary.LittleEndian.PutUint32(b[4:], randPart)
	return base64.RawURLEncoding.EncodeToString(b)
}
