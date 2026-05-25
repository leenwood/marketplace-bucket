package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/marketplace/marketplace-bucket/internal/core/domain"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Error("cannot connect to redis", "addr", addr, "error", err)
		os.Exit(1)
	}

	carts := seedCarts()

	for _, cart := range carts {
		data, err := json.Marshal(cart)
		if err != nil {
			log.Error("marshal cart", "user_id", cart.UserID, "error", err)
			os.Exit(1)
		}
		if err := client.Set(ctx, "cart:"+cart.UserID, data, 7*24*time.Hour).Err(); err != nil {
			log.Error("save cart", "user_id", cart.UserID, "error", err)
			os.Exit(1)
		}
		log.Info("seeded cart",
			"user_id", cart.UserID,
			"items", len(cart.Items),
			"total", fmt.Sprintf("%.2f", cart.Total()),
		)
	}

	log.Info("done", "carts_seeded", len(carts))
}

func seedCarts() []*domain.Cart {
	return []*domain.Cart{
		cartAlice(),
		cartBob(),
		cartCharlie(),
	}
}

func cartAlice() *domain.Cart {
	c := domain.NewCart("user-alice")
	c.AddItem(&domain.Item{ProductID: "prod-001", Name: "Wireless Headphones", Price: 59.99, Quantity: 1})
	c.AddItem(&domain.Item{ProductID: "prod-002", Name: "USB-C Hub", Price: 34.50, Quantity: 2})
	c.AddItem(&domain.Item{ProductID: "prod-003", Name: "Laptop Stand", Price: 27.00, Quantity: 1})
	return c
}

func cartBob() *domain.Cart {
	c := domain.NewCart("user-bob")
	c.AddItem(&domain.Item{ProductID: "prod-004", Name: "Mechanical Keyboard", Price: 129.99, Quantity: 1})
	c.AddItem(&domain.Item{ProductID: "prod-005", Name: "Mouse Pad XL", Price: 19.99, Quantity: 1})
	return c
}

func cartCharlie() *domain.Cart {
	c := domain.NewCart("user-charlie")
	c.AddItem(&domain.Item{ProductID: "prod-006", Name: "Webcam 1080p", Price: 79.00, Quantity: 1})
	c.AddItem(&domain.Item{ProductID: "prod-007", Name: "Ring Light", Price: 45.00, Quantity: 2})
	c.AddItem(&domain.Item{ProductID: "prod-008", Name: "HDMI Cable 2m", Price: 12.99, Quantity: 3})
	c.AddItem(&domain.Item{ProductID: "prod-009", Name: "Desk Organizer", Price: 22.50, Quantity: 1})
	return c
}
