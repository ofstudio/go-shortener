package ipcheck

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// defaultRemoteIPHeader - заголовок по умолчанию для определения IP
const defaultRemoteIPHeader = "X-Real-IP"

// Whitelist - реализация Provider, которая проверяет,
// что IP-адрес клиента присутствует в списке разрешенных адресов или сетей.
// IP-адрес клиента определяется по http-заголовку / grpc-метаданным (X-Real-IP по умолчанию).
type Whitelist struct {
	allowedIPNets   []*net.IPNet
	allowedIPs      []net.IP
	remoteIPHeaders []string
}

// NewWhitelist - конструктор Whitelist.
// addrs - список разрешенных адресов в формате IPv4 или CIDR.
func NewWhitelist(addrs ...string) *Whitelist {
	allowedIPNets := make([]*net.IPNet, 0, len(addrs))
	allowedIPs := make([]net.IP, 0, len(addrs))

	for _, addr := range addrs {
		// Пытаемся разобрать адрес как CIDR
		_, ipNet, err := net.ParseCIDR(addr)
		// Если не получилось, то пытаемся разобрать как IP
		if err != nil {
			ip := net.ParseIP(addr)
			if ip != nil {
				// Если получилось как IP, то добавляем в список разрешенных IP
				allowedIPs = append(allowedIPs, ip)
			}
		} else {
			// Если получилось как CIDR, то добавляем в список разрешенных сетей
			allowedIPNets = append(allowedIPNets, ipNet)
		}
	}

	return &Whitelist{
		allowedIPNets:   allowedIPNets,
		allowedIPs:      allowedIPs,
		remoteIPHeaders: []string{defaultRemoteIPHeader},
	}
}

// UseHeaders - устанавливает список http-заголовков и grpc-метаданных,
// из которых будет определяться реальный IP клиента.
func (wl *Whitelist) UseHeaders(headers ...string) *Whitelist {
	wl.remoteIPHeaders = headers
	return wl
}

// Handler - HTTP-middleware для проверки IP-адреса клиента.
func (wl *Whitelist) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем IP клиента
		ip := wl.httpRemoteIP(r)
		// Проверяем, что IP клиента найден в заголовках и что он разрешен
		if ip == nil || !wl.IsAllowed(ip) {
			// Если не найден или не разрешен, то возвращаем ошибку 403
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		// Если разрешен, то вызываем следующий обработчик
		next.ServeHTTP(w, r)
	})
}

// Interceptor - GRPC-middleware для проверки IP-адреса клиента.
func (wl *Whitelist) Interceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "metadata not found")
	}
	// Получаем IP клиента
	ip := wl.grpcRemoteIP(md)
	// Проверяем, что IP клиента найден в заголовках и что он разрешен
	if ip == nil || !wl.IsAllowed(ip) {
		// Если не найден или не разрешен, то возвращаем ошибку 403
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	// Если разрешен, то вызываем следующий обработчик
	return handler(ctx, req)
}

// IsAllowed - проверяет, что IP-адрес клиента присутствует в списке разрешенных адресов или сетей.
func (wl *Whitelist) IsAllowed(ip net.IP) bool {
	// Проверяем, что IP клиента присутствует в списке разрешенных сетей
	for _, allowedIPNet := range wl.allowedIPNets {
		if allowedIPNet.Contains(ip) {
			return true
		}
	}
	// Проверяем, что IP клиента присутствует в списке разрешенных IP
	for _, allowedIP := range wl.allowedIPs {
		if ip.Equal(allowedIP) {
			return true
		}
	}
	return false
}

// httpRemoteIP - возвращает IP-адрес клиента из http-заголовка.
// Если IP-адрес не найден в заголовках, то возвращается nil.
func (wl *Whitelist) httpRemoteIP(r *http.Request) net.IP {
	for _, header := range wl.remoteIPHeaders {
		ip := net.ParseIP(r.Header.Get(header))
		if ip != nil {
			return ip
		}
	}
	return nil
}

// grpcRemoteIP - возвращает IP-адрес клиента из метаданных GRPC-запроса.
// Если IP-адрес не найден в метаданных, то возвращается nil.
func (wl *Whitelist) grpcRemoteIP(md metadata.MD) net.IP {

	for _, key := range wl.remoteIPHeaders {
		val := md.Get(key)
		if len(val) == 0 {
			continue
		}
		ip := net.ParseIP(val[0])
		if ip != nil {
			return ip
		}
	}
	return nil
}
