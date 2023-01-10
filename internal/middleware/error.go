package middleware

import (
	"errors"
	"net/http"
)

var (
	// ErrSigningError - ошибка при подписании токена
	ErrSigningError = errors.New("signing error")
	// ErrInvalidToken - невалидный токен
	ErrInvalidToken = errors.New("invalid token")
)

// respondWithError - возвращает клиенту http-ошибку, соответствующую ошибке middleware
func respondWithError(w http.ResponseWriter, _ error) {
	switch {
	// tbd...
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
