package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app"
	"io"
	"log"
	"net/http"
)

// URLsResource - хандлеры для сокращения URL
type URLsResource struct {
	app *app.App
}

func NewURLsResource(app *app.App) *URLsResource {
	return &URLsResource{app}
}

func (rs URLsResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", rs.Get)
	r.Post("/", rs.Post)
	return r
}

// Get - эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL
// и возвращает ответ с кодом http.StatusTemporaryRedirect (307) и оригинальным URL
// в HTTP-заголовке Location.
func (rs URLsResource) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fullURL, err := rs.app.GetLongURL(id)
	if err != nil {
		rs.error(w, err)
		return
	}
	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

// Post - эндпоинт POST / принимает в теле запроса строку URL для сокращения
// и возвращает ответ http.StatusCreated (201) и сокращённым URL
// в виде текстовой строки в теле.
func (rs URLsResource) Post(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		rs.error(w, err)
		return
	}

	if len(b) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortURL, err := rs.app.CreateShortURL(string(b))
	if err != nil {
		rs.error(w, err)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(shortURL)); err != nil {
		log.Println(err)
	}

}

// error - возвращает http-ошибку, соотвествующую ошибке приложения
func (rs URLsResource) error(w http.ResponseWriter, err error) {
	if errors.Is(err, app.ErrURLNotFound) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, app.ErrValidation) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
