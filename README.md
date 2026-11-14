# 🛡️ Go High-Performance Rate Limiter & Telemetry Middleware

Middleware HTTP de **Rate Limiting** em **Go (Golang)** baseado no algoritmo **Token Bucket**, integrado a um motor nativo de **Observabilidade e Telemetria de Hardware (CPU, Goroutines, Heap Go e RAM Linux)**.

## 🚀 Destaques de Engenharia

- **Algoritmo Token Bucket Thread-Safe:** Reposição contínua de tokens baseada em timestamps (`time.Now()`), protegido por `sync.Mutex` de baixíssima contenção.
- **Telemetria em Tempo Real (Zero Dependências Externas):** Coleta métricas do runtime Go (`runtime.MemStats`, Goroutines, GC) e lê diretamente `/proc/meminfo` e `/proc/loadavg` do kernel Linux.
- **Dupla Política de Limitação (IP & API Key):** Limites diferenciados para tráfego anônimo (por IP) e autenticado via cabeçalho `X-API-Key`.
- **Prevenção de Memory Leaks (TTL Eviction):** Goroutine em background que limpa clientes inativos após o tempo de expiração.
- **Cabeçalhos HTTP Padrão:** Informa `X-RateLimit-Limit`, `X-RateLimit-Remaining` e `Retry-After` (HTTP 429).

## 📊 Exemplo de Resposta Real (`GET /api/v1/data`)

    {
      "message": "Requisição processada dentro do limite de taxa (Token Bucket)!",
      "status": "success",
      "telemetry": {
        "uptime": "1m5s",
        "cpu": {
          "logical_cores": 12,
          "active_goroutines": 4,
          "load_avg_1m": 0.04,
          "load_avg_5m": 0.09,
          "estimated_cpu_usage_percent": "0.33%"
        },
        "go_process_memory": {
          "heap_alloc_mb": 0.29,
          "total_alloc_mb": 0.29,
          "sys_reserved_mb": 6.58,
          "gc_cycles": 0
        },
        "host_system_ram": {
          "total_mb": 7520,
          "used_mb": 1060,
          "available_mb": 6460,
          "used_percent": 14.09
        }
      },
      "timestamp": "2026-11-14T22:50:26Z"
    }

## 🛠️ Como Executar os Testes, Benchmarks e Servidor

    go test -v -race -bench=. ./...
    go run ./cmd/server
