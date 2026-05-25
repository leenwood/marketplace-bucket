package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/marketplace/marketplace-bucket/internal/config"
	"github.com/marketplace/marketplace-bucket/internal/app/service"
	"github.com/marketplace/marketplace-bucket/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString("load config: " + err.Error() + "\n")
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := service.RunServer(ctx, cfg, log); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
