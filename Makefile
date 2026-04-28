BIN     := ./bin/marketplace-bucket
MAIN    := ./cmd/server
LDFLAGS := -ldflags="-s -w"

.PHONY: build test lint run tidy docker-up docker-down

## build: compile the binary to ./bin/marketplace-bucket
build:
	go build $(LDFLAGS) -o $(BIN) $(MAIN)

## test: run all tests with race detector and coverage
test:
	go test ./... -v -race -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

## test-short: run only unit tests (no integration)
test-short:
	go test ./... -short -race

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## run: run the service locally (requires Redis on localhost:6379)
run:
	go run $(MAIN)

## tidy: tidy and verify modules
tidy:
	go mod tidy
	go mod verify

## docker-up: start all services (marketplace-bucket + Redis + Jaeger)
docker-up:
	docker-compose up -d --build

## docker-down: stop and remove all containers
docker-down:
	docker-compose down

## help: print this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
