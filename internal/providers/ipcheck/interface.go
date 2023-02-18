package ipcheck

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

// Provider - провайдер проверки IP-адреса клиента.
// Реализует middleware для http и grpc.
type Provider interface {
	// IsAllowed - проверяет, разрешен ли IP-адрес
	IsAllowed(ip net.IP) bool
	// Handler - http.HandlerFunc
	Handler(next http.Handler) http.Handler
	// Interceptor - grpc.UnaryServerInterceptor
	Interceptor(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error)
}
