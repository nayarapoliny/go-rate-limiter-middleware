# 🛡️ Go High-Performance Rate Limiter Middleware

Middleware HTTP de **Rate Limiting** em **Go (Golang)** baseado no algoritmo **Token Bucket**, projetado para proteger APIs e microsserviços contra sobrecarga, *brute-force* e picos de tráfego (DDoS).

## 🚀 Destaques de Engenharia

- **Algoritmo Token Bucket Thread-Safe:** Cálculo preciso de reposição contínua de tokens baseado em *timestamps* (`time.Now()`), protegido por `sync.Mutex` de baixíssima contenção.
- **Dupla Política de Limitação (IP & API Key):** Aplica limites diferenciados para tráfego anônimo (por IP extraído de `X-Forwarded-For` ou `RemoteAddr`) e tráfego autenticado via cabeçalho `X-API-Key`.
- **Prevenção de Memory Leaks (TTL Eviction):** Goroutine em *background* que remove automaticamente baldes de clientes inativos após o tempo de expiração configurável.
- **Cabeçalhos HTTP Padrão:** Informa ao cliente em todas as respostas sua cota total (`X-RateLimit-Limit`), saldo restante (`X-RateLimit-Remaining`) e tempo de espera em caso de bloqueio HTTP 429 (`Retry-After`).
- **Benchmark Concorrente Integrado:** Suíte de testes unitários e benchmark paralelo (`b.RunParallel`) validando alta vazão sob múltiplas goroutines.

## 🛠️ Como Executar os Testes e Benchmarks

    go test -v -race -bench=. ./...

## ⚡ Como Executar o Servidor e Testar o Bloqueio (HTTP 429)

    go run ./cmd/server

**Disparar 7 requisições seguidas para testar o bloqueio na 6ª requisição:**

    for i in {1..7}; do curl -i -s http://localhost:8082/api/v1/data | head -n 1; done

**Testar cota prioritária enviando `X-API-Key`:**

    curl -i -H "X-API-Key: vip-token-123" http://localhost:8082/api/v1/data
