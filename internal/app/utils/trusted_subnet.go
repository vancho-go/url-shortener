package utils

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware выполняет проверку вхождения X-Real-IP из запроса в доверенную подсеть
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get("X-Real-IP")
			if !isIPTrusted(realIP, trustedSubnet) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// isIPTrusted проверяет IP на вхождение в доверенную подсеть.
func isIPTrusted(ip string, trustedSubnet string) bool {
	if trustedSubnet == "" {
		return false
	}
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return false
	}
	if !subnet.Contains(net.ParseIP(ip)) {
		return false
	}

	return true
}
