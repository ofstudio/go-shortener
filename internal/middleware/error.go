package middleware

import (
	"errors"
	"net/http"
)

var (
	ErrSigningError = errors.New("signing error") // ErrSigningError - ошибка при подписании токена
	ErrInvalidToken = errors.New("invalid token") // ErrInvalidToken - невалидный токен
)

// respondWithError - возвращает клиенту http-ошибку, соответствующую ошибке middleware
func respondWithError(w http.ResponseWriter, _ error) {
	switch {
	// tbd...
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
