# marketplace-bucket

[🇷🇺 Русский](README.ru.md)

A production-ready Go microservice for shopping cart management backed by Redis.
Built to demonstrate real-world patterns in clean architecture, observability, and distributed systems.

---

## What This Repository Demonstrates

- **Clean architecture** — four strict layers (domain → port → usecase → infra/app) enforced by `go-arch-lint`
- **OpenTelemetry tracing** — distributed trace propagation across HTTP handlers and Redis operations
- **Prometheus metrics** — RED metrics and cart operation counters via an isolated registry
- **Structured logging** — JSON logs with `request_id` and `trace_id` on every line
- **Redis-backed storage** — cart persisted as JSON blob with configurable TTL (default 7 days)
- **Graceful shutdown** — in-flight requests drained before Redis and OTel flush

---

## Architecture Overview

```
cmd/server/main.go
        │
        ▼
internal/app/service/server.go      ← wires all deps, starts HTTP, handles shutdown
        │
        ▼
internal/app/http/server.go         ← NewServer: routes + middleware chain
        │
        ├── middleware: otelhttp → Recover → Logger → RequestID → MaxBodySize
        │
        ├── POST   /api/v1/cart/{userID}/items
        ├── GET    /api/v1/cart/{userID}
        ├── PATCH  /api/v1/cart/{userID}/items/{productID}
        ├── DELETE /api/v1/cart/{userID}/items/{productID}
        ├── DELETE /api/v1/cart/{userID}
        │
        ├── GET /health   GET /ready   GET /metrics
        └── GET /swagger/   GET /debug/pprof/*

        │
        ▼
internal/core/usecase/cart.go       ← CartUseCase: business logic
        │
        ▼
internal/infra/storage/redis/       ← CartRepository: Redis key cart:{userID}
```

---

## Key Design Decisions

| Area | Decision |
|---|---|
| Storage | Redis key `cart:{user_id}` as JSON blob; 7-day TTL |
| Transport | HTTP only (REST/JSON) |
| Config | Pure env-var helpers — no Viper or other third-party config lib |
| Observability | Prometheus at `/metrics` (isolated registry), OTel OTLP HTTP or stdout |
| Arch enforcement | `go-arch-lint` fails the build if layer boundaries are violated |
| Tests | Unit tests mock repo via interface; Redis integration tests use `miniredis` |

---

## Observability

### Logs
All logs are JSON. Every request line includes:
```json
{"time":"...","level":"INFO","msg":"request","method":"POST","path":"/api/v1/cart/u1/items",
 "status":200,"latency":"1.2ms","request_id":"uuid","trace_id":"otel-trace-id"}
```

### Traces
OpenTelemetry spans cover:
- HTTP handler (root span via `otelhttp`)
- Redis Get / Set / Del operations

### Metrics
| Metric | Type | Labels |
|---|---|---|
| `platform_http_requests_total` | Counter | method, path, status |
| `platform_http_request_duration_seconds` | Histogram | method, path |
| `cart_operations_total` | Counter | operation |

---

## Local Setup

**Prerequisites:** Docker, Docker Compose, Go 1.22+, [go-task](https://taskfile.dev/installation/)

```bash
# Install go-task (once)
go install github.com/go-task/task/v3/cmd/task@latest

# 1. Clone and enter
git clone https://github.com/leenwood/marketplace-bucket
cd marketplace-bucket

# 2. Start infrastructure (Redis + Jaeger)
task docker:up

# 3. Run the server
task run
```

Services available locally:

| Service | URL |
|---|---|
| API server | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Prometheus metrics | http://localhost:8080/metrics |
| Jaeger UI | http://localhost:16686 |

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HTTP_ADDR` | `:8080` | HTTP listen address |
| `HTTP_READ_TIMEOUT` | `15s` | Read timeout |
| `HTTP_WRITE_TIMEOUT` | `15s` | Write timeout |
| `HTTP_PPROF_ENABLED` | `false` | Enable `/debug/pprof/*` |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | `` | Redis password |
| `REDIS_DB` | `0` | Redis DB index |
| `CART_TTL` | `168h` | Cart expiry TTL (7 days) |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log format (json/text) |
| `OTEL_ENABLED` | `false` | Enable OpenTelemetry |
| `OTEL_EXPORTER` | `stdout` | Exporter type (stdout/otlp) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4318` | OTLP HTTP endpoint |
| `OTEL_SERVICE_NAME` | `marketplace-bucket` | Service name in traces |

---

## API Examples

```bash
# Health check
curl http://localhost:8080/health

# Readiness probe (pings Redis)
curl http://localhost:8080/ready

# Add item to cart
curl -X POST http://localhost:8080/api/v1/cart/user-123/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "prod-42",
    "name": "Wireless Headphones",
    "price": 59.99,
    "quantity": 2
  }'

# Get cart
curl http://localhost:8080/api/v1/cart/user-123

# Update item quantity
curl -X PATCH http://localhost:8080/api/v1/cart/user-123/items/prod-42 \
  -H "Content-Type: application/json" \
  -d '{"quantity": 3}'

# Remove item
curl -X DELETE http://localhost:8080/api/v1/cart/user-123/items/prod-42

# Clear entire cart
curl -X DELETE http://localhost:8080/api/v1/cart/user-123
```

---

## Task Commands

Run `task --list` to see all available commands.

| Command | Description |
|---|---|
| `task build` | Compile binary to `./bin/marketplace-bucket` |
| `task run` | Run HTTP server locally |
| `task test` | Run all tests with race detector |
| `task test:cover` | Tests with coverage report |
| `task test:integration` | Run integration tests (requires Docker) |
| `task lint` | Run golangci-lint |
| `task arch` | Check architecture constraints via go-arch-lint |
| `task vet` | Run go vet |
| `task fmt` | Format code with gofmt |
| `task docker:up` | Start Redis + Jaeger containers |
| `task docker:down` | Stop containers |
| `task docker:reset` | Stop containers and wipe volumes |
| `task docs` | Regenerate Swagger docs |

---

## Testing

Unit tests cover the `CartUseCase` with a mock repository (interface-driven), and the HTTP handlers via `httptest`.

Integration tests use `miniredis` to run a real in-process Redis — no Docker required for `task test`.

```bash
task test
task test:integration
```

---

## Possible Production Improvements

- **Rate limiting** — per-user request limits to protect Redis from burst writes
- **Distributed TTL management** — separate job to clean up expired carts and emit metrics
- **Optimistic locking** — Redis `WATCH`/`MULTI`/`EXEC` for concurrent cart updates
- **Event publishing** — publish `cart.updated` events to Kafka for downstream order services
- **Auth middleware** — JWT validation to bind `userID` from token claims, not URL path
- **Alerting rules** — Prometheus alerts for high error rate, Redis latency spikes
