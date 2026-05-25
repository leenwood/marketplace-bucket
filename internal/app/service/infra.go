package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/marketplace/marketplace-bucket/internal"
	goredis "github.com/redis/go-redis/v9"

	"github.com/marketplace/marketplace-bucket/internal/platform/metrics"
	"github.com/marketplace/marketplace-bucket/internal/platform/tracing"
)

type Infra struct {
	Cfg     *internal.Config
	Log     *slog.Logger
	Metrics *metrics.Metrics
	Redis   *goredis.Client

	shutdownTracing tracing.ShutdownFunc
}

func initInfra(ctx context.Context, cfg *internal.Config, log *slog.Logger) (*Infra, error) {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	shutdownTracing, err := tracing.Init(initCtx, tracing.Config{
		Enabled:     cfg.OTel.Enabled,
		Exporter:    cfg.OTel.Exporter,
		Endpoint:    cfg.OTel.Endpoint,
		ServiceName: cfg.OTel.ServiceName,
	})
	if err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	m := metrics.New()

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(initCtx).Err(); err != nil {
		_ = shutdownTracing(initCtx)
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return &Infra{
		Cfg:             cfg,
		Log:             log,
		Metrics:         m,
		Redis:           redisClient,
		shutdownTracing: shutdownTracing,
	}, nil
}

func (i *Infra) Shutdown(ctx context.Context) {
	if err := i.Redis.Close(); err != nil {
		i.Log.Error("close redis", "error", err)
	}
	if err := i.shutdownTracing(ctx); err != nil {
		i.Log.Error("flush traces", "error", err)
	}
}
