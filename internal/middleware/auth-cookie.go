package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/models"
)

type contextKey struct {
	name string
}

var UserIDCtxKey = &contextKey{"user_id"}

const (
	authCookieName = "auth_token"
	lenData        = 8 // uint64 size
	lenSignature   = sha256.Size
	lenToken       = lenData + lenSignature
)

// AuthCookie - middleware для проверки и установки аутентификационной куки.
// Выдает пользователю симметрично подписанную куку,
// содержащую уникальный идентификатор пользователя, если такой куки не существует
// или она не проходит проверку подлинности.
type AuthCookie struct {
	srv    *services.Container
	secret []byte
	domain string
	maxAge int
	secure bool
}

func NewAuthCookie(srv *services.Container) *AuthCookie {
	return &AuthCookie{srv: srv}
}

// WithSecret - устанавливает секрет для подписи куки
func (m *AuthCookie) WithSecret(secret []byte) *AuthCookie {
	m.secret = secret
	return m
}

// WithDomain - устанавливает домен для куки
func (m *AuthCookie) WithDomain(host string) *AuthCookie {
	domain, _, _ := net.SplitHostPort(host)
	m.domain = domain
	return m
}

// WithTTL - устанавливает время жизни куки
func (m *AuthCookie) WithTTL(ttl time.Duration) *AuthCookie {
	m.maxAge = int(ttl / time.Second)
	return m
}

// WithSecure - устанавливает значение secure для куки
func (m *AuthCookie) WithSecure(secure bool) *AuthCookie {
	m.secure = secure
	return m
}

// Handler - возвращает middleware
func (m *AuthCookie) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authCookieName)
		// Если кука не найдена - устанавливаем куку и передаем обработку запроса дальше
		if errors.Is(err, http.ErrNoCookie) || cookie == nil {
			m.setCookie(w, r, next)
			return
		}

		// Проверяем подпись куки
		userID, err := m.verifyToken(cookie.Value)
		// Если подпись не совпадает - устанавливаем куку и передаем обработку запроса дальше
		if errors.Is(err, ErrInvalidToken) {
			m.setCookie(w, r, next)
			return
		} else if err != nil {
			respondWithError(w, err)
			return
		}

		// Устанавливаем userID в контекст запроса и передаем запрос дальше
		next.ServeHTTP(w, m.withContext(r, userID))
	})
}

// setCookie - устанавливает куку и передает запрос дальше
func (m *AuthCookie) setCookie(w http.ResponseWriter, r *http.Request, next http.Handler) {
	user := &models.User{}
	if err := m.srv.UserService.Create(r.Context(), user); err != nil {
		respondWithError(w, err)
		return
	}
	token, err := m.createToken(user.ID)
	if err != nil {
		respondWithError(w, err)
		return
	}
	cookie := &http.Cookie{
		Name:     authCookieName,
		Value:    token,
		Domain:   m.domain,
		Path:     "/",
		MaxAge:   m.maxAge,
		Secure:   m.secure,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}

	http.SetCookie(w, cookie)
	next.ServeHTTP(w, m.withContext(r, user.ID))
}

// withContext - устанавливает userID в контекст запроса
func (m *AuthCookie) withContext(r *http.Request, id uint) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), UserIDCtxKey, id))
}

// createToken - создает и подписывает куку для пользователя
func (m *AuthCookie) createToken(userID uint) (string, error) {
	b := make([]byte, lenData)
	binary.BigEndian.PutUint64(b, uint64(userID))
	signature, err := m.sign(b, m.secret)
	if err != nil {
		return "", ErrSigningError
	}
	return base64.RawURLEncoding.EncodeToString(append(b, signature...)), nil
}

// verifyToken - проверяет подпись куки
func (m *AuthCookie) verifyToken(tokenStr string) (uint, error) {
	if tokenStr == "" {
		return 0, ErrInvalidToken
	}
	token, err := base64.RawURLEncoding.DecodeString(tokenStr)
	if err != nil || len(token) < lenToken {
		return 0, ErrInvalidToken
	}
	refSignature, err := m.sign(token[:lenData], m.secret)
	if err != nil {
		return 0, ErrSigningError
	}
	if !hmac.Equal(token[lenData:], refSignature) {
		return 0, ErrInvalidToken
	}
	userID := binary.BigEndian.Uint64(token[:lenData])

	return uint(userID), nil
}

// sign - подписывает данные: hmac/sha256
func (m *AuthCookie) sign(data, secret []byte) ([]byte, error) {
	h := hmac.New(sha256.New, secret)
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// UserIDFromContext - возвращает userID из контекста
func UserIDFromContext(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(UserIDCtxKey).(uint)
	if !ok {
		return 0, false
	}
	return id, true
}
