# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`marketplace-bucket` — production-ready Go microservice for shopping cart management backed by Redis.

## Commands

```bash
task build          # compile binary to ./bin/marketplace-bucket
task run            # run the HTTP server locally (requires Redis on :6379)
task test           # run all tests with race detector
task test:cover     # tests + coverage report
task lint           # run golangci-lint via go tool
task arch           # check architecture via go-arch-lint
task docker:up      # start Redis + Jaeger
task docker:down    # stop containers
```

Single test:
```bash
go test ./internal/... -run TestName -v -race
```

## Architecture

Four-layer clean architecture enforced by `go-arch-lint`:

```
cmd/server/main.go                  # entry point — signal.NotifyContext, RunServer

internal/
  config/                           # Foundation: env-var config, no third-party libs
  core/
    domain/                         # Foundation: Cart, Item, domain errors
    port/                           # Contracts: CartRepository, CartService interfaces
    usecase/                        # Use Cases: CartUseCase (→ port, domain, logger)
  infra/
    storage/redis/                  # Infrastructure: Redis-backed CartRepository
  platform/
    logger/                         # Foundation: slog constructor + context helpers
    metrics/                        # Foundation: isolated Prometheus registry
    tracing/                        # Foundation: OTel init (OTLP HTTP or stdout)
  app/
    http/
      handler/                      # Application: HTTP handlers (→ port, metrics, logger)
      middleware/                   # Application: Chain, Recover, Logger, RequestID, MaxBodySize
      server.go                     # Application: NewServer wiring + all standard routes
    service/
      infra.go                      # Bootstrap: Infra struct, initInfra, Infra.Shutdown
      server.go                     # Bootstrap: RunServer (wires deps, starts HTTP, graceful shutdown)
```

## Key Design Decisions

- **Storage**: Redis key `cart:{user_id}` as JSON blob; 7-day TTL.
- **Transport**: HTTP only (REST/JSON). gRPC removed in favour of base-infra standard.
- **Config**: Pure env-var helpers — no Viper or other third-party config lib.
- **Observability**: Prometheus at `/metrics` (isolated registry), OTel tracing (OTLP HTTP or stdout).
- **Middleware chain**: otelhttp → Recover → Logger (injects trace_id) → RequestID → MaxBodySize.
- **Standard routes**: `/health`, `/ready` (pings Redis), `/metrics`, `/swagger/`, `/debug/pprof/*`.
- **Graceful shutdown**: signal.NotifyContext + 30 s shutdown window; Redis closed before OTel flush.
- **Tests**: unit tests mock repo via interface; Redis integration tests use miniredis (not containers).
- **Tooling**: go-task (Taskfile.yml); golangci-lint as `go tool` dep in go.mod.
