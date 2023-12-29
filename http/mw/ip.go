package mw

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

// RealIPHandler gets the real IP address behind a proxy
func RealIPHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, err := getIP(r)
			if err == nil {
				log := zerolog.Ctx(r.Context())
				log.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str(fieldKey, ip)
				})
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-Ip")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	ips := r.Header.Get("X-Forwarded-For")
	for _, ip := range strings.Split(ips, ",") {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	return "", fmt.Errorf("no valid IP found")
}
