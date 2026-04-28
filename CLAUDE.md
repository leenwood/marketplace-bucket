# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a `marketplace-bucket` — a production-ready Go microservice for shopping cart management. It has not been implemented yet; `promt.md` contains the full specification.

## Commands

Once the project is scaffolded (see `promt.md`), the intended Makefile targets are:

```bash
make build   # compile the service
make test    # run unit tests
make lint    # run golangci-lint
make run     # run the service locally
```

Run a single test:
```bash
go test ./internal/... -run TestName -v
```

Start infrastructure (Redis):
```bash
docker-compose up -d
```

## Architecture

Clean architecture with three layers:

```
cmd/server/main.go          # entry point, wiring, graceful shutdown
internal/
  handler/                  # transport layer — REST (HTTP/JSON) + gRPC
  service/                  # business logic, interfaces for all dependencies
  repository/               # Redis storage (Hash + JSON, TTL 7 days)
pkg/                        # reusable packages (config via Viper, metrics, tracing)
proto/                      # .proto definitions for gRPC
```

Dependency flow: `handler → service → repository`. Each layer depends only on interfaces, not concrete types.

## Key Design Decisions

- **Storage**: Redis key `cart:{user_id}` as Hash or JSON; 7-day TTL per cart.
- **Transport**: Both REST and gRPC exposed simultaneously.
- **Config**: Environment variables via Viper.
- **Observability**: Prometheus metrics at `/metrics`, OpenTelemetry tracing.
- **Go version**: 1.22+. No global state. All dependencies injected via interfaces.
- **gRPC errors**: Return proper gRPC status codes alongside HTTP status codes on the REST side.
