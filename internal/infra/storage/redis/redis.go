package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/marketplace/marketplace-bucket/internal/core/domain"
)

const (
	keyPrefix  = "cart:"
	defaultTTL = 7 * 24 * time.Hour
)

type Repository struct {
	client goredis.Cmdable
	ttl    time.Duration
	tracer trace.Tracer
}

func New(client goredis.Cmdable, ttl time.Duration) *Repository {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Repository{
		client: client,
		ttl:    ttl,
		tracer: otel.Tracer("cart/infra/storage/redis"),
	}
}

func cartKey(userID string) string {
	return keyPrefix + userID
}

func (r *Repository) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	ctx, span := r.tracer.Start(ctx, "redis.Get")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	data, err := r.client.Get(ctx, cartKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, domain.ErrCartNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var cart domain.Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("unmarshal cart: %w", err)
	}

	return &cart, nil
}

func (r *Repository) Save(ctx context.Context, cart *domain.Cart) error {
	ctx, span := r.tracer.Start(ctx, "redis.Save")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", cart.UserID))

	data, err := json.Marshal(cart)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("marshal cart: %w", err)
	}

	if err := r.client.Set(ctx, cartKey(cart.UserID), data, r.ttl).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID string) error {
	ctx, span := r.tracer.Start(ctx, "redis.Delete")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	if err := r.client.Del(ctx, cartKey(userID)).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}
