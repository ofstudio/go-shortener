package models

// ShortURL - модель сокращенной ссылки
type ShortURL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	UserID      uint   `json:"user_id"`
}
