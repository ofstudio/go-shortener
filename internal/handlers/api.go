package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"mime"
	"net/http"
)

// APIHandlers - HTTP хандлеры JSON API для сервиса services.ShortenerService
type APIHandlers struct {
	srv *services.ShortenerService
}

func NewAPIHandlers(srv *services.ShortenerService) *APIHandlers {
	return &APIHandlers{srv: srv}
}

func (h APIHandlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/shorten", h.CreateShortURL)
	return r
}

// createShortURLReq - запрос для создания сокращенного URL
type createShortURLReq struct {
	URL string `json:"url"`
}

// createShortURLRes - ответ на запрос CreateShortURL
type createShortURLRes struct {
	Result string `json:"result"`
}

// CreateShortURL - принимает в теле запроса строку URL для сокращения:
//    {"url":"<url>"}
//
// Возвращает ответ http.StatusCreated (201) и сокращенный URL в виде JSON:
//    {"result":"<shorten_url>"}
func (h APIHandlers) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	reqBody := createShortURLReq{}
	if err := parseJSONRequest(r, &reqBody); err != nil {
		respondWithError(w, err)
	}

	shortURL, err := h.srv.CreateShortURL(reqBody.URL)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, createShortURLRes{Result: shortURL})
}

// parseJSON - парсит запрос в теле запроса в структуру.
// Проверяет наличие заголовка Content-Type: application/json
func parseJSONRequest(r *http.Request, v interface{}) error {
	// check content type is JSON
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		return ErrValidation
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return ErrValidation
	}
	return nil
}

// respondWithJSON - HTTP ответ в формате JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
