package handlers

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ofstudio/go-shortener/internal/pkgerrors"
	"github.com/ofstudio/go-shortener/internal/providers/auth"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// @Title Go-Shortener API
// @Description API для сокращения ссылок
// @Version 1.0
// @Contact.name Oleg Fomin
// @Contact.email ofstudio@yandex.ru
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @In cookie
// @Name auth_token

// APIHandlers - HTTP-хендлеры для JSON API
type APIHandlers struct {
	u *usecases.Container
}

// NewAPIHandlers - конструктор APIHandlers
func NewAPIHandlers(u *usecases.Container) *APIHandlers {
	return &APIHandlers{u}
}

// PublicRoutes - возвращает роутер с хендлерами
func (h APIHandlers) PublicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/shorten", h.shortURLCreate)
	r.Post("/shorten/batch", h.shortURLCreateBatch)
	r.Get("/user/urls", h.shortURLGetByUserID)
	r.Delete("/user/urls", h.shortURLDeleteBatch)
	return r
}

// InternalRoutes - возвращает роутер с внутренними хендлерами
func (h APIHandlers) InternalRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/stats", h.stats)
	return r
}

// shortURLCreate - принимает в теле запроса строку URL для сокращения:
//
//	{"url":"<url>"}
//
// Возвращает ответ http.StatusCreated (201) и сокращенный URL в виде JSON:
//
//	{"result":"<shorten_url>"}
//
// @Tags shorten
// @Summary Создает сокращенную ссылку
// @Security cookieAuth
// @ID shortURLCreate
// @Accept  json
// @Produce json
// @Param   request body handlers.shortURLCreate.reqType true "Запрос"
// @Success 201 {object} handlers.shortURLCreate.resType
// @Failure 400
// @Failure 401
// @Failure 409 {object} handlers.shortURLCreate.resType
// @Failure 410
// @Failure 500
// @Router /shorten [post]
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
	userID, ok := auth.FromContext(r.Context())
	if !ok {
		respondWithError(w, pkgerrors.ErrAuth)
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
	shortURL, err := h.u.ShortURL.Create(r.Context(), userID, reqJSON.URL)

	if err != nil && !errors.Is(err, pkgerrors.ErrDuplicate) {
		respondWithError(w, err)
		return
	}

	// Если ссылка уже существует возвращаем код 409
	if errors.Is(err, pkgerrors.ErrDuplicate) {
		statusCode = http.StatusConflict
	}

	// Возвращаем ответ
	respondWithJSON(w,
		statusCode,
		resType{Result: h.u.ShortURL.Resolve(shortURL.ID)})
}

// shortURLCreateBatch - принимает в теле запроса список строк URL для сокращения:
//
//	[
//	    {
//	        "correlation_id": "<строковый идентификатор>",
//	        "original_url": "<URL для сокращения>"
//	    },
//	    ...
//	]
//
// Возвращает ответ http.StatusCreated (201) и сокращенный URL в виде JSON:
//
//	[
//	    {
//	        "correlation_id": "<строковый идентификатор из объекта запроса>",
//	        "short_url": "<shorten_url>"
//	    },
//	    ...
//	]
//
// @Tags shorten
// @Summary Создает несколько сокращенных ссылок
// @Security cookieAuth
// @ID shortURLCreateBatch
// @Accept  json
// @Produce json
// @Param   request body handlers.shortURLCreateBatch.reqType true "Запрос"
// @Success 201 {array} handlers.shortURLCreateBatch.resType
// @Failure 400 {string} Bad Request
// @Failure 401 {string} string "Unauthorized"
// @Failure 410 {string} string "Gone"
// @Failure 500 {string} string "Internal server Error"
// @Router /shorten/batch [post]
func (h APIHandlers) shortURLCreateBatch(w http.ResponseWriter, r *http.Request) {
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
	userID, ok := auth.FromContext(r.Context())
	if !ok {
		respondWithError(w, pkgerrors.ErrAuth)
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
		shortURL, err := h.u.ShortURL.Create(r.Context(), userID, item.OriginalURL)
		if err != nil && !errors.Is(err, pkgerrors.ErrDuplicate) {
			respondWithError(w, err)
			return
		}
		resJSON[i] = resType{
			CorrelationID: item.CorrelationID,
			ShortURL:      h.u.ShortURL.Resolve(shortURL.ID),
		}
	}

	// Возвращаем ответ
	respondWithJSON(w, http.StatusCreated, resJSON)
}

// shortURLDeleteBatch - помечает ссылки пользователя как удаленные.
// Формат запроса:
//
//	[ "a", "b", "c", "d", ...]
//
// Возвращает ответ http.StatusAccepted (202)
//
// @Tags user
// @Summary Удаляет несколько сокращенных ссылок
// @Security cookieAuth
// @ID shortURLDeleteBatch
// @Accept  json
// @Produce json
// @Param   request body handlers.shortURLDeleteBatch.reqType true "Запрос"
// @Success 202
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /user/urls [delete]
func (h APIHandlers) shortURLDeleteBatch(w http.ResponseWriter, r *http.Request) {
	// Структура элемента запроса
	type reqType []string

	// Проверяем аутентифицирован ли пользователь
	userID, ok := auth.FromContext(r.Context())
	if !ok {
		respondWithError(w, pkgerrors.ErrAuth)
		return
	}

	// Читаем body запроса
	reqJSON := make(reqType, 0)
	if err := parseJSONRequest(r, &reqJSON); err != nil {
		respondWithError(w, err)
		return
	}

	if len(reqJSON) == 0 {
		respondWithError(w, pkgerrors.ErrValidation)
		return
	}
	// Отправляем ответ
	w.WriteHeader(http.StatusAccepted)

	// Удаляем ссылки
	_ = h.u.ShortURL.DeleteBatch(r.Context(), userID, reqJSON)
}

// shortURLGetByUserID - возвращает список сокращенных ссылок пользователя.
// Формат ответа:
//
//	[
//	    {
//	        "short_url": "http://...",
//	        "original_url": "http://..."
//	    },
//	    ...
//	]
//
// @Tags user
// @Summary Возвращает список сокращенных ссылок пользователя
// @Security cookieAuth
// @ID shortURLGetByUserID
// @Produce json
// @Success 200 {array} handlers.shortURLGetByUserID.resType
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /user/urls [get]
func (h APIHandlers) shortURLGetByUserID(w http.ResponseWriter, r *http.Request) {
	// Структура ответа
	type resType struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	// Проверяем аутентифицирован ли пользователь
	userID, ok := auth.FromContext(r.Context())
	if !ok {
		respondWithError(w, pkgerrors.ErrAuth)
		return
	}

	// Получаем список сокращенных ссылок пользователя
	shortURLs, err := h.u.ShortURL.GetByUserID(r.Context(), userID)
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
			ShortURL:    h.u.ShortURL.Resolve(shortURLs[i].ID),
			OriginalURL: shortURLs[i].OriginalURL,
		}
	}

	// Возвращаем ответ
	respondWithJSON(w, http.StatusOK, res)
}

// stats - возвращает статистику сервиса.
// Формат ответа:
//
//	{
//	    "urls": 100,
//	    "users": 10
//	}
//
// @Tags internal
// @Summary Возвращает статистику сервиса
// @Security ipAuth
// @ID stats
// @Produce json
// @Success 200 {object} handlers.stats.resType
// @Failure 403
// @Failure 500
// @Router /internal/stats [get]
func (h APIHandlers) stats(w http.ResponseWriter, r *http.Request) {
	// Структура ответа
	type resType struct {
		Users int `json:"users"`
		URLs  int `json:"urls"`
	}

	// Получаем статистику
	userCount, err := h.u.User.Count(r.Context())
	if err != nil {
		respondWithError(w, err)
		return
	}
	shortURLCount, err := h.u.ShortURL.Count(r.Context())
	if err != nil {
		respondWithError(w, err)
		return
	}

	// Формируем ответ
	res := resType{
		Users: userCount,
		URLs:  shortURLCount,
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
		return pkgerrors.ErrValidation
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return pkgerrors.ErrValidation
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
