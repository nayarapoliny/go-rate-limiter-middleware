package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware_IPAndAPITokenLimits(t *testing.T) {
	limiter := NewLimiter(Config{
		IPRatePerSec:     1,
		IPBurst:          2,
		TokenRatePerSec:  10,
		TokenBurst:       5,
		CleanupInterval:  time.Minute,
		ClientExpiration: time.Minute,
	})
	defer limiter.Stop()

	handler := Middleware(limiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Deve permitir as 2 primeiras requisições por IP (burst = 2) e bloquear a 3ª com HTTP 429
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
		req.RemoteAddr = "192.168.1.50:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if i <= 2 && rec.Code != http.StatusOK {
			t.Fatalf("requisição %d deveria passar (200), mas retornou %d", i, rec.Code)
		}
		if i == 3 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("requisição 3 deveria ser bloqueada (429), mas retornou %d", rec.Code)
		}
	}

	// 2. Com X-API-Key VIP, o mesmo cliente tem cota separada e maior (burst = 5)
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
		req.RemoteAddr = "192.168.1.50:1234"
		req.Header.Set("X-API-Key", "vip-secret-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("requisição autenticada %d deveria passar (200), obteve %d", i, rec.Code)
		}
	}
}

func BenchmarkLimiter_AllowConcurrent(b *testing.B) {
	limiter := NewLimiter(Config{
		IPRatePerSec: 100000,
		IPBurst:      100000,
	})
	defer limiter.Stop()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow("ip:10.0.0.1", false)
		}
	})
}
