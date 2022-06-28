package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"mime"
	"net/http"
)

// APIHandlers - HTTP-хендлеры для JSON API
type APIHandlers struct {
	srv *services.Services
}

func NewAPIHandlers(srv *services.Services) *APIHandlers {
	return &APIHandlers{srv}
}

func (h APIHandlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/shorten", h.createShortURL)
	r.Get("/user/urls", h.getUserURLs)
	return r
}

// createShortURL - принимает в теле запроса строку URL для сокращения:
//    {"url":"<url>"}
//
// Возвращает ответ http.StatusCreated (201) и сокращенный URL в виде JSON:
//    {"result":"<shorten_url>"}
func (h APIHandlers) createShortURL(w http.ResponseWriter, r *http.Request) {
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
	reqBody := reqType{}
	if err := parseJSONRequest(r, &reqBody); err != nil {
		respondWithError(w, err)
	}

	// Создаем сокращенную ссылку
	shortURL, err := h.srv.ShortURLService.Create(r.Context(), userID, reqBody.URL)
	if err != nil {
		respondWithError(w, err)
		return
	}

	// Возвращаем ответ
	respondWithJSON(w,
		http.StatusCreated,
		resType{Result: h.srv.ShortURLService.Resolve(shortURL.ID)})
}

// getUserURLs - возвращает список сокращенных ссылок пользователя.
// Формат ответа:
//    [
//        {
//            "short_url": "http://...",
//            "original_url": "http://..."
//        },
//        ...
//    ]
func (h APIHandlers) getUserURLs(w http.ResponseWriter, r *http.Request) {
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
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
