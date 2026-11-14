package ratelimiter

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func Middleware(l *Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
		var clientKey string
		var isToken bool

		if apiKey != "" {
			clientKey = "token:" + apiKey
			isToken = true
		} else {
			ip := extractClientIP(r)
			clientKey = "ip:" + ip
			isToken = false
		}

		allowed, limit, remaining, retryAfter := l.Allow(clientKey, isToken)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":               "rate limit exceeded: too many requests",
				"retry_after_seconds": retryAfter,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
