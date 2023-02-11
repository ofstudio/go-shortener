package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ofstudio/go-shortener/internal/app"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// HTTPHandlers - HTTP-хендлеры приложения
type HTTPHandlers struct {
	u *usecases.Container
}

// NewHTTPHandlers - конструктор HTTPHandlers
func NewHTTPHandlers(srv *usecases.Container) *HTTPHandlers {
	return &HTTPHandlers{u: srv}
}

// Routes - возвращает роутер для HTTP-хендлеров
func (h HTTPHandlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/ping", h.ping)
	r.Get("/{id}", h.shortURLRedirectToOriginal)
	r.Post("/", h.shortURLCreate)
	return r
}

// shortURLRedirectToOriginal - принимает в качестве URL-параметра идентификатор сокращённого URL
// и возвращает ответ с кодом http.StatusTemporaryRedirect (307) и оригинальным URL
// в HTTP-заголовке Location.
func (h HTTPHandlers) shortURLRedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	shortURL, err := h.u.ShortURL.GetByID(r.Context(), id)
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
		respondWithError(w, app.ErrAuth)
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

	originalURL := string(b)
	statusCode := http.StatusCreated
	shortURL, err := h.u.ShortURL.Create(r.Context(), userID, originalURL)

	// Если ссылка уже существует, возвращаем её
	if errors.Is(err, app.ErrDuplicate) {
		statusCode = http.StatusConflict
		shortURL, err = h.u.ShortURL.GetByOriginalURL(r.Context(), originalURL)
	}

	if err != nil {
		respondWithError(w, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(h.u.ShortURL.Resolve(shortURL.ID)))
}

// ping - вызывает Health.Check.
// Возвращает ответ http.StatusOK (200) или http.StatusInternalServerError (500).
func (h HTTPHandlers) ping(w http.ResponseWriter, r *http.Request) {
	err := h.u.Health.Check(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
