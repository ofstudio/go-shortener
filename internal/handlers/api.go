package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"mime"
	"net/http"
)

// APIHandlers - HTTP-хендлеры для JSON API
type APIHandlers struct {
	srv *services.Container
}

func NewAPIHandlers(srv *services.Container) *APIHandlers {
	return &APIHandlers{srv}
}

func (h APIHandlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/shorten", h.shortURLCreate)
	r.Post("/shorten/batch", h.shortURLBatchCreate)
	r.Get("/user/urls", h.shortURLGetByUserID)
	return r
}

// shortURLCreate - принимает в теле запроса строку URL для сокращения:
//    {"url":"<url>"}
//
// Возвращает ответ http.StatusCreated (201) и сокращенный URL в виде JSON:
//    {"result":"<shorten_url>"}
func (h APIHandlers) shortURLCreate(w http.ResponseWriter, r *http.Request) {
	// Структура запроса
	type reqType struct {
		URL string `json:"url"`
	}
	// Структура ответа
	type resType struct {
		Result string `json:"result"`
	}

	// Проверяем аутентифицирован ли пользователь
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, ErrAuth)
		return
	}

	// Читаем body запроса
	reqJSON := reqType{}
	if err := parseJSONRequest(r, &reqJSON); err != nil {
		respondWithError(w, err)
		return
	}

	// Создаем сокращенную ссылку
	statusCode := http.StatusCreated
	shortURL, err := h.srv.ShortURLService.Create(r.Context(), userID, reqJSON.URL)

	// Если ссылка уже существует, запрашиваем ее
	if errors.Is(err, services.ErrDuplicate) {
		statusCode = http.StatusConflict
		shortURL, err = h.srv.ShortURLService.GetByOriginalURL(r.Context(), reqJSON.URL)
	}

	if err != nil {
		respondWithError(w, err)
		return
	}

	// Возвращаем ответ
	respondWithJSON(w,
		statusCode,
		resType{Result: h.srv.ShortURLService.Resolve(shortURL.ID)})
}

// shortURLBatchCreate - принимает в теле запроса список строк URL для сокращения:
//    [
//        {
//            "correlation_id": "<строковый идентификатор>",
//            "original_url": "<URL для сокращения>"
//        },
//        ...
//    ]
// Возвращает ответ http.StatusCreated (201) и сокращенный URL в виде JSON:
//    [
//        {
//            "correlation_id": "<строковый идентификатор из объекта запроса>",
//            "short_url": "<shorten_url>"
//        },
//        ...
//    ]
func (h APIHandlers) shortURLBatchCreate(w http.ResponseWriter, r *http.Request) {
	// Структура элемента запроса
	type reqType struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	// Структура элемента ответа
	type resType struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	// Проверяем аутентифицирован ли пользователь
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, ErrAuth)
		return
	}

	// Читаем body запроса
	reqJSON := make([]reqType, 0)
	if err := parseJSONRequest(r, &reqJSON); err != nil {
		respondWithError(w, err)
		return
	}

	// Создаем сокращенные ссылки
	resJSON := make([]resType, len(reqJSON))
	for i, item := range reqJSON {
		shortURL, err := h.srv.ShortURLService.Create(r.Context(), userID, item.OriginalURL)
		if err != nil {
			respondWithError(w, err)
			return
		}
		resJSON[i] = resType{
			CorrelationID: item.CorrelationID,
			ShortURL:      h.srv.ShortURLService.Resolve(shortURL.ID),
		}
	}

	// Возвращаем ответ
	respondWithJSON(w, http.StatusCreated, resJSON)
}

// shortURLGetByUserID - возвращает список сокращенных ссылок пользователя.
// Формат ответа:
//    [
//        {
//            "short_url": "http://...",
//            "original_url": "http://..."
//        },
//        ...
//    ]
func (h APIHandlers) shortURLGetByUserID(w http.ResponseWriter, r *http.Request) {
	// Структура ответа
	type resType struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	// Проверяем аутентифицирован ли пользователь
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, ErrAuth)
		return
	}

	// Получаем список сокращенных ссылок пользователя
	shortURLs, err := h.srv.ShortURLService.GetByUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, err)
		return
	}

	// Если пользователь не имеет сокращенных ссылок, возвращаем 204 No Content
	if len(shortURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Формируем ответ
	res := make([]resType, len(shortURLs))
	for i := range shortURLs {
		res[i] = resType{
			ShortURL:    h.srv.ShortURLService.Resolve(shortURLs[i].ID),
			OriginalURL: shortURLs[i].OriginalURL,
		}
	}

	// Возвращаем ответ
	respondWithJSON(w, http.StatusOK, res)
}

// parseJSONRequest - парсит запрос в теле запроса в структуру.
// Проверяет наличие заголовка Content-Type: application/json
func parseJSONRequest(r *http.Request, v interface{}) error {
	// Проверяем наличие Content-Type: application/json
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
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
