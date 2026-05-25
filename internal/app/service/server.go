package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	apphttp "github.com/marketplace/marketplace-bucket/internal/app/http"
	"github.com/marketplace/marketplace-bucket/internal/app/http/handler"
	"github.com/marketplace/marketplace-bucket/internal/config"
	"github.com/marketplace/marketplace-bucket/internal/core/usecase"
	redisstorage "github.com/marketplace/marketplace-bucket/internal/infra/storage/redis"
)

func RunServer(ctx context.Context, cfg *config.Config, log *slog.Logger) error {
	infra, err := initInfra(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("init infra: %w", err)
	}

	repo := redisstorage.New(infra.Redis, cfg.App.CartTTL)
	uc := usecase.NewCart(repo, infra.Log)
	h := handler.New(uc, infra.Metrics, infra.Log)

	srv := apphttp.NewServer(cfg.HTTP, apphttp.Deps{
		Log:     infra.Log,
		Metrics: infra.Metrics,
		Handler: h,
		Ping: func() error {
			return infra.Redis.Ping(ctx).Err()
		},
	})

	srvErr := make(chan error, 1)
	go func() {
		log.Info("HTTP server starting", "addr", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	select {
	case err := <-srvErr:
		infra.Shutdown(ctx)
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP shutdown", "error", err)
	}

	infra.Shutdown(shutdownCtx)
	log.Info("shutdown complete")
	return nil
}
