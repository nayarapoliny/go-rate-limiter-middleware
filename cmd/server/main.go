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
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Requisição processada dentro do limite de taxa (Token Bucket)!",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	protectedHandler := ratelimiter.Middleware(limiter, mux)

	fmt.Printf("🛡️ Servidor Go Rate Limiter Middleware ativo em http://localhost:%s/api/v1/data\n", port)
	fmt.Println("   • Limite por IP: Burst de 5 reqs (recarga de 2 req/s)")
	fmt.Println("   • Limite por X-API-Key: Burst de 20 reqs (recarga de 10 req/s)")
	log.Fatal(http.ListenAndServe(":"+port, protectedHandler))
}
