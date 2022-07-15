package models

// URLMaxLen - максимальная длина исходного URL в байтах.
// Формально, размер URL ничем не ограничен.
// Разные версии разных браузеров имеют свои конкретные ограничения: от 2048 байт до нескольких мегабайт.
// В случае нашего сервиса необходимо некое разумное ограничение.
const URLMaxLen = 4096

// ShortURL - модель сокращенной ссылки
type ShortURL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	UserID      uint   `json:"user_id"`
}
