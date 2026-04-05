FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cart-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cart-service ./cmd/server

# ────────────────────────────────────────────────────────────────────────────

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /cart-service /cart-service

EXPOSE 8080 9090

ENTRYPOINT ["/cart-service"]
