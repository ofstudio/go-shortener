package middleware

import (
	"errors"
	"net/http"
)

// ErrSigningError - ошибка при подписании токена
var ErrSigningError = errors.New("signing error")

// ErrInvalidToken - невалидный токен
var ErrInvalidToken = errors.New("invalid token")

// respondWithError - возвращает клиенту http-ошибку, соответствующую ошибке middleware
func respondWithError(w http.ResponseWriter, _ error) {
	switch {
	// tbd...
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
