// Package application wires all service dependencies and manages server lifecycle.
package application

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpchandler "github.com/marketplace/marketplace-bucket/internal/handler/grpc"
	httphandler "github.com/marketplace/marketplace-bucket/internal/handler/http"
	"github.com/marketplace/marketplace-bucket/internal/repository"
	"github.com/marketplace/marketplace-bucket/internal/service"
	"github.com/marketplace/marketplace-bucket/pkg/config"
	"github.com/marketplace/marketplace-bucket/pkg/metrics"
	"github.com/marketplace/marketplace-bucket/pkg/pb"
	"github.com/marketplace/marketplace-bucket/pkg/tracing"
)

// App holds the wired application and owns its lifecycle.
type App struct {
	cfg *config.Config
}

// New creates an App from the given config.
func New(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

// Run initialises all dependencies, starts HTTP and gRPC servers, and blocks
// until ctx is cancelled. It performs a graceful shutdown before returning.
func (a *App) Run(ctx context.Context) error {
	cfg := a.cfg

	// ── Tracing ───────────────────────────────────────────────────────────────

	shutdownTracing, err := tracing.Init(ctx, cfg.Otel.Endpoint, cfg.Otel.ServiceName)
	if err != nil {
		slog.Warn("tracing unavailable, continuing without it", "error", err)
		shutdownTracing = func(context.Context) error { return nil }
	}
	defer func() {
		tctx, tcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer tcancel()
		if err := shutdownTracing(tctx); err != nil {
			slog.Error("shutdown tracing", "error", err)
		}
	}()

	// ── Infrastructure ────────────────────────────────────────────────────────

	m := metrics.New("cart")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return &startupError{"connect to redis", err}
	}
	defer redisClient.Close()

	// ── Service layer ─────────────────────────────────────────────────────────

	repo := repository.NewRedisRepository(redisClient, cfg.CartTTL)
	svc := service.NewCartService(repo, m)

	// ── HTTP server ───────────────────────────────────────────────────────────

	mux := http.NewServeMux()
	httphandler.NewHandler(svc, m).RegisterRoutes(mux)
	mux.Handle("GET /metrics", promhttp.HandlerFor(m.Registry(), promhttp.HandlerOpts{}))

	httpServer := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		slog.Info("HTTP server listening", "addr", cfg.HTTP.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server", "error", err)
		}
	}()

	// ── gRPC server ───────────────────────────────────────────────────────────

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcMetricsInterceptor(m)),
	)
	pb.RegisterCartServiceServer(grpcServer, grpchandler.NewHandler(svc, m))
	reflection.Register(grpcServer)

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPC.Addr)
		if err != nil {
			slog.Error("gRPC listen", "addr", cfg.GRPC.Addr, "error", err)
			return
		}
		slog.Info("gRPC server listening", "addr", cfg.GRPC.Addr)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server", "error", err)
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────

	<-ctx.Done()
	slog.Info("shutdown signal received")

	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown", "error", err)
	}

	slog.Info("all servers stopped")
	return nil
}

// grpcMetricsInterceptor records request count and latency for every unary RPC.
func grpcMetricsInterceptor(m *metrics.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		statusLabel := "ok"
		if err != nil {
			statusLabel = "error"
		}
		m.GRPCRequestDuration.WithLabelValues(info.FullMethod, statusLabel).Observe(time.Since(start).Seconds())
		m.GRPCRequestsTotal.WithLabelValues(info.FullMethod, statusLabel).Inc()
		return resp, err
	}
}

// startupError wraps an infrastructure initialisation failure.
type startupError struct {
	op  string
	err error
}

func (e *startupError) Error() string { return e.op + ": " + e.err.Error() }
func (e *startupError) Unwrap() error { return e.err }
