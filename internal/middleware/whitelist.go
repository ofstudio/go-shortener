package middleware

import (
	"net"
	"net/http"
)

// defaultRealIPHeader - заголовок по умолчанию для определения реального IP
const defaultRealIPHeader = "X-Real-IP"

// Whitelist - middleware для проверки IP-адреса клиента.
// Проверяет, что IP-адрес клиента присутствует в списке разрешенных адресов или сетей.
// Если IP-адрес не разрешен, то возвращает ошибку 403.
// IP-адрес клиента определяется по http-заголовку (X-Real-IP по умолчанию).
type Whitelist struct {
	allowedIPNets []*net.IPNet
	allowedIPs    []net.IP
	realIPHeaders []string
}

// NewWhitelist - конструктор Whitelist.
// addrs - список разрешенных адресов в формате IPv4 или CIDR.
func NewWhitelist(addrs ...string) *Whitelist {
	allowedIPNets := make([]*net.IPNet, 0, len(addrs))
	allowedIPs := make([]net.IP, 0, len(addrs))

	for _, addr := range addrs {
		// Пытаемся разобрать адрес как CIDR
		ip, ipNet, err := net.ParseCIDR(addr)
		// Если не получилось, то пытаемся разобрать как IP
		if err != nil {
			ip = net.ParseIP(addr)
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
		allowedIPNets: allowedIPNets,
		allowedIPs:    allowedIPs,
		realIPHeaders: []string{defaultRealIPHeader},
	}
}

// UseHeaders - устанавливает список http-заголовков, из которых будет определяться реальный IP клиента.
func (m *Whitelist) UseHeaders(headers ...string) *Whitelist {
	m.realIPHeaders = headers
	return m
}

// Handler - middleware для проверки IP-адреса клиента.
func (m *Whitelist) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем IP клиента
		ip := getRealIP(r, m.realIPHeaders)
		// Проверяем, что IP клиента найден в заголовках и что он разрешен
		if ip == nil || !m.isAllowedIP(ip) {
			// Если не найден или не разрешен, то возвращаем ошибку 403
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		// Если разрешен, то вызываем следующий middleware
		next.ServeHTTP(w, r)
	})
}

// isAllowedIP - проверяет, что IP-адрес клиента присутствует в списке разрешенных адресов или сетей.
func (m *Whitelist) isAllowedIP(ip net.IP) bool {
	// Проверяем, что IP клиента присутствует в списке разрешенных сетей
	for _, allowedIPNet := range m.allowedIPNets {
		if allowedIPNet.Contains(ip) {
			return true
		}
	}
	// Проверяем, что IP клиента присутствует в списке разрешенных IP
	for _, allowedIP := range m.allowedIPs {
		if ip.Equal(allowedIP) {
			return true
		}
	}
	return false
}

// getRealIP - возвращает IP-адрес клиента из http-заголовка.
// Если IP-адрес не найден в заголовках, то возвращается nil.
func getRealIP(r *http.Request, headers []string) net.IP {
	for _, header := range headers {
		ip := net.ParseIP(r.Header.Get(header))
		if ip != nil {
			return ip
		}
	}
	return nil
}
