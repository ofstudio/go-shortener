package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"io"
	"net/http"
)

// HTTPHandlers - HTTP-хендлеры для сервиса services.ShortURLService
type HTTPHandlers struct {
	shortURLService *services.ShortURLService
}

func NewHTTPHandlers(srv *services.ShortURLService) *HTTPHandlers {
	return &HTTPHandlers{srv}
}

func (h HTTPHandlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", h.shortURLRedirectToOriginal)
	r.Post("/", h.shortURLCreate)
	return r
}

// shortURLRedirectToOriginal - принимает в качестве URL-параметра идентификатор сокращённого URL
// и возвращает ответ с кодом http.StatusTemporaryRedirect (307) и оригинальным URL
// в HTTP-заголовке Location.
func (h HTTPHandlers) shortURLRedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	shortURL, err := h.shortURLService.GetByID(id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	http.Redirect(w, r, shortURL.OriginalURL, http.StatusTemporaryRedirect)
}

// shortURLCreate - принимает в теле запроса строку URL для сокращения
// и возвращает ответ http.StatusCreated (201) и сокращённым URL
// в виде текстовой строки в теле.
func (h HTTPHandlers) shortURLCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, ErrAuth)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, err)
		return
	}

	if len(b) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortURL, err := h.shortURLService.Create(userID, string(b))
	if err != nil {
		respondWithError(w, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(h.shortURLService.Resolve(shortURL.ID)))
}
