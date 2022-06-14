package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"io"
	"net/http"
)

// ShortenerHandlers - HTTP-хендлеры для сервиса services.ShortenerService
type ShortenerHandlers struct {
	srv *services.ShortenerService
}

func NewShortenerHandlers(srv *services.ShortenerService) *ShortenerHandlers {
	return &ShortenerHandlers{srv}
}

func (h ShortenerHandlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", h.GetLongURL)
	r.Post("/", h.CreateShortURL)
	return r
}

// GetLongURL - принимает в качестве URL-параметра идентификатор сокращённого URL
// и возвращает ответ с кодом http.StatusTemporaryRedirect (307) и оригинальным URL
// в HTTP-заголовке Location.
func (h ShortenerHandlers) GetLongURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fullURL, err := h.srv.GetLongURL(id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

// CreateShortURL - принимает в теле запроса строку URL для сокращения
// и возвращает ответ http.StatusCreated (201) и сокращённым URL
// в виде текстовой строки в теле.
func (h ShortenerHandlers) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, err)
		return
	}

	if len(b) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortURL, err := h.srv.CreateShortURL(string(b))
	if err != nil {
		respondWithError(w, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortURL))
}
