package handlers

import (
	"errors"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"net/http"
)

var ErrValidation = errors.New("validation error")

// respondWithError - возвращает клиенту http-ошибку, соотвествующую ошибке сервиса
func respondWithError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrShortURLNotFound):
		http.Error(w, "Not found", http.StatusNotFound)
	case errors.Is(err, services.ErrValidation),
		errors.Is(err, ErrValidation):
		http.Error(w, "Bad request", http.StatusBadRequest)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
