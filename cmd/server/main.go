package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nayarapoliny/go-rate-limiter-middleware/internal/ratelimiter"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	limiter := ratelimiter.NewLimiter(ratelimiter.Config{
		IPRatePerSec:     2,
		IPBurst:          5,
		TokenRatePerSec:  10,
		TokenBurst:       20,
		CleanupInterval:  1 * time.Minute,
		ClientExpiration: 5 * time.Minute,
	})
	defer limiter.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(map[string]interface{}{
			"status":    "success",
			"message":   "Requisição processada dentro do limite de taxa (Token Bucket)!",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"telemetry": ratelimiter.CollectTelemetry(),
		})
	})

	protectedHandler := ratelimiter.Middleware(limiter, mux)

	fmt.Printf("🛡️ Servidor Go Rate Limiter + Telemetria ativo em http://localhost:%s/api/v1/data\n", port)
	log.Fatal(http.ListenAndServe(":"+port, protectedHandler))
}
