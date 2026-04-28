package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/marketplace/marketplace-bucket/internal/application"
	"github.com/marketplace/marketplace-bucket/pkg/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
	}()

	if err := application.New(cfg).Run(ctx); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}
