package handlers

import (
	"errors"
	"net/http"

	"github.com/ofstudio/go-shortener/internal/app/services"
)

var (
	ErrValidation = errors.New("validation error")
	ErrAuth       = errors.New("unauthorized")
)

// respondWithError - возвращает клиенту http-ошибку, соответствующую ошибке сервиса
func respondWithError(w http.ResponseWriter, err error) {
	switch {
	// 404
	case errors.Is(err, services.ErrNotFound):
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	// 400
	case errors.Is(err, services.ErrValidation),
		errors.Is(err, ErrValidation):
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	// 401
	case errors.Is(err, ErrAuth):
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 410
	case errors.Is(err, services.ErrDeleted):
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
	// 500
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
