package handlers

import (
	"errors"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"net/http"
)

// errorResponse - возвращает клиенту http-ошибку, соотвествующую ошибке сервиса
func errorResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrShortURLNotFound):
		http.Error(w, "Not found", http.StatusNotFound)
	case errors.Is(err, services.ErrValidation):
		http.Error(w, "Bad request", http.StatusBadRequest)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
