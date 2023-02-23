package auth

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

// Provider - провайдер проверки токена пользователя.
// Реализует middleware для http и grpc.
type Provider interface {
	// CreateToken - создает токен для пользователя.
	CreateToken(userID uint) (string, error)
	// VerifyToken - проверяет токен пользователя.
	VerifyToken(token string) (uint, error)
	// Handler - http.HandlerFunc
	Handler(next http.Handler) http.Handler
	// Interceptor - grpc.UnaryServerInterceptor
	Interceptor(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error)
}
