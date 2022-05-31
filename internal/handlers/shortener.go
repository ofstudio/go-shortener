package handlers

import (
	"errors"
	"github.com/ofstudio/go-shortener/internal/app"
	"io"
	"log"
	"net/http"
)

// Shortener - http.Handler для сокращения URL
type Shortener struct {
	app *app.App
}

func NewShortener(app *app.App) *Shortener {
	return &Shortener{app}
}

func (h Shortener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPost:
		h.post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// get - эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL
// и возвращает ответ с кодом http.StatusTemporaryRedirect (307) и оригинальным URL
// в HTTP-заголовке Location.
func (h Shortener) get(w http.ResponseWriter, r *http.Request) {
	fullURL, err := h.app.GetLongURL(r.URL.Path[1:])
	if err != nil {
		h.error(w, err)
		return
	}
	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

// post - эндпоинт POST / принимает в теле запроса строку URL для сокращения
// и возвращает ответ http.StatusCreated (201) и сокращённым URL
// в виде текстовой строки в теле.
func (h Shortener) post(w http.ResponseWriter, r *http.Request) {

	if r.RequestURI != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		h.error(w, err)
		return
	}

	if len(b) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortURL, err := h.app.CreateShortURL(string(b))
	if err != nil {
		h.error(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(shortURL)); err != nil {
		log.Println(err)
	}

}

// error - возвращает http-ошибку, соотвествующую ошибке приложения
func (h Shortener) error(w http.ResponseWriter, err error) {
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
