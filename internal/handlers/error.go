package handlers

import (
	"errors"
	"net/http"

	"github.com/ofstudio/go-shortener/internal/usecases"
)

// ErrValidation - ошибка валидации
var ErrValidation = errors.New("validation error")

// ErrAuth - ошибка авторизации
var ErrAuth = errors.New("unauthorized")

// respondWithError - возвращает клиенту http-ошибку, соответствующую ошибке сервиса
func respondWithError(w http.ResponseWriter, err error) {
	switch {
	// 404
	case errors.Is(err, usecases.ErrNotFound):
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	// 400
	case errors.Is(err, usecases.ErrValidation),
		errors.Is(err, ErrValidation):
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	// 401
	case errors.Is(err, ErrAuth):
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 410
	case errors.Is(err, usecases.ErrDeleted):
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
	// 500
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
