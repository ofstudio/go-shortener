package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

const (
	httpCookieName  = "auth_token"
	grpcMetadataKey = httpCookieName
)

const (
	lenUserID    = 8 // uint64 size
	lenExpiresAt = 8 // uint64 size
	lenData      = lenUserID + lenExpiresAt
	lenSignature = sha256.Size
)

// CookieOpts - опции для HTTP-куки.
type CookieOpts struct {
	Domain   string
	Path     string
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

// SHA256Provider - провайдер аутентификации с использованием HMAC-SHA256 для подписи токена.
// Предоставляет методы для создания и проверки токенов.
// Токен состоит из 3 частей:
//  1. id пользователя (8 байт)
//  2. Таймстамп окончания жизни токена (8 байт, unix-время в секундах)
//  3. Подпись (32 байта, подпись HMAC-SHA256)
//
// Токен закодирован в строку в формате base64 (RFC 4648)
// и может быть передан в заголовке Authorization или в куках.
//
// Если в запросе передан невалидный или просроченный токен,
// то создается новый пользователь и клиенту возвращается новый токен.
type SHA256Provider struct {
	u          *usecases.User
	CookieOpts *CookieOpts   // Опции для HTTP-куки
	secret     []byte        // Секретный ключ для подписи токена
	ttl        time.Duration // Время жизни токена
}

// NewSHA256Provider - конструктор SHA256Provider.
func NewSHA256Provider(cfg *config.Config, u *usecases.User) *SHA256Provider {
	return &SHA256Provider{
		u:      u,
		secret: []byte(cfg.AuthSecret),
		ttl:    cfg.AuthTTL,
		CookieOpts: &CookieOpts{
			Domain:   cfg.BaseURL.Hostname(),
			Path:     "/",
			Secure:   cfg.BaseURL.Scheme == "https",
			HttpOnly: true,
			SameSite: http.SameSiteDefaultMode,
		},
	}
}

// CreateToken - создает токен для пользователя.
func (p *SHA256Provider) CreateToken(userID uint) (string, error) {
	tokenBytes := make([]byte, lenData)
	expiresAt := time.Now().Add(p.ttl).Unix()
	binary.BigEndian.PutUint64(tokenBytes, uint64(userID))
	binary.BigEndian.PutUint64(tokenBytes[lenUserID:], uint64(expiresAt))
	signature, err := p.sign(tokenBytes)
	if err != nil {
		return "", ErrSigningError
	}
	return base64.RawURLEncoding.EncodeToString(append(tokenBytes, signature...)), nil
}

// VerifyToken - проверяет токен и возвращает id пользователя.
func (p *SHA256Provider) VerifyToken(token string) (uint, error) {
	if token == "" {
		return 0, ErrInvalidToken
	}
	// Декодируем токен
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, ErrInvalidToken
	}
	// Проверяем подпись токена
	if err = p.verify(tokenBytes); err != nil {
		return 0, err
	}
	// Проверяем, что токен не просрочен.
	expiresAt := int64(binary.BigEndian.Uint64(tokenBytes[lenUserID:lenData]))
	if expiresAt < time.Now().Unix() {
		return 0, ErrExpiredToken
	}
	// Возвращаем id пользователя
	userID := uint(binary.BigEndian.Uint64(tokenBytes[:lenUserID]))
	return userID, nil
}

// Handler - HTTP-middleware для проверки авторизации
func (p *SHA256Provider) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие токена в http-куке кго и валидность
		if cookie, err := r.Cookie(httpCookieName); err == nil && cookie != nil {
			if userID, err := p.VerifyToken(cookie.Value); err == nil {
				// Если найден валидный токен - устанавливаем userID в контекст
				// и передаем запрос дальше
				next.ServeHTTP(w, r.WithContext(ToContext(r.Context(), userID)))
				return
			}
		}

		// Иначе - создаем нового пользователя и токен
		userID, token, err := p.newUserWithToken(r.Context())
		if err != nil {
			log.Err(err).Msg("auth: failed to create new user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем токен в http-куку
		http.SetCookie(w, &http.Cookie{
			Name:     httpCookieName,
			Value:    token,
			Domain:   p.CookieOpts.Domain,
			Path:     p.CookieOpts.Path,
			MaxAge:   int(p.ttl / time.Second),
			Secure:   p.CookieOpts.Secure,
			HttpOnly: p.CookieOpts.HttpOnly,
			SameSite: http.SameSiteDefaultMode,
		})
		// Устанавливаем userID в контекст и передаем запрос дальше
		next.ServeHTTP(w, r.WithContext(ToContext(r.Context(), userID)))
	})
}

// Interceptor - GRPC-middleware для проверки авторизации
func (p *SHA256Provider) Interceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Проверяем наличие токена и валидность в заголовке запроса
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if token := md.Get(grpcMetadataKey); len(token) > 0 {
			if userID, err := p.VerifyToken(token[0]); err == nil {
				// Если найден валидный токен - устанавливаем userID в контекст
				// и передаем запрос дальше
				return handler(ToContext(ctx, userID), req)
			}
		}
	}

	// Иначе - создаем нового пользователя и токен
	userID, token, err := p.newUserWithToken(ctx)
	if err != nil {
		log.Err(err).Msg("auth: failed to create new user")
		return nil, status.Error(codes.Internal, "internal error")
	}
	// Устанавливаем токен в заголовок ответа
	if err = grpc.SetHeader(ctx, metadata.Pairs(grpcMetadataKey, token)); err != nil {
		log.Err(err).Msg("auth: failed to set header")
		return nil, status.Error(codes.Internal, "internal error")
	}
	// Устанавливаем userID в контекст и передаем запрос дальше
	return handler(ToContext(ctx, userID), req)
}

// newUserWithToken - создает нового пользователя и возвращает его id и токен.
func (p *SHA256Provider) newUserWithToken(ctx context.Context) (uint, string, error) {
	user := &models.User{}
	if err := p.u.Create(ctx, user); err != nil {
		return 0, "", err
	}
	token, err := p.CreateToken(user.ID)
	if err != nil {
		return 0, "", err
	}
	return user.ID, token, nil
}

// sign - подписывает данные: hmac/sha256
func (p *SHA256Provider) sign(tokenBytes []byte) ([]byte, error) {
	h := hmac.New(sha256.New, p.secret)
	if _, err := h.Write(tokenBytes); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// verify - проверяет подпись данных: hmac/sha256
func (p *SHA256Provider) verify(tokenBytes []byte) error {
	if len(tokenBytes) != lenData+lenSignature {
		return ErrInvalidToken
	}
	refSignature, err := p.sign(tokenBytes[:lenData])
	if err != nil {
		return ErrSigningError
	}
	if !hmac.Equal(tokenBytes[lenData:], refSignature) {
		return ErrInvalidToken
	}
	return nil
}
