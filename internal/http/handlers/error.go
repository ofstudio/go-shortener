package handlers

import (
	"net/http"

	"github.com/ofstudio/go-shortener/internal/pkgerrors"
)

// respondWithError - возвращает клиенту http-ошибку, соответствующую ошибке приложения
func respondWithError(w http.ResponseWriter, err error) {
	if appError, ok := err.(*pkgerrors.Error); ok {
		http.Error(w, http.StatusText(appError.HTTPStatus), appError.HTTPStatus)
	} else {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
