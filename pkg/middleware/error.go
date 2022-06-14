package middleware

import (
	"net/http"
)

// respondWithError - возвращает клиенту http-ошибку, соотвествующую ошибке middleware
func respondWithError(w http.ResponseWriter, _ error) {
	switch {
	// tbd...
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}