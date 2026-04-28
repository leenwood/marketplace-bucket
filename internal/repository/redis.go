package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/marketplace/marketplace-bucket/internal/domain"
)

const (
	keyPrefix  = "cart:"
	defaultTTL = 7 * 24 * time.Hour
)

// RedisRepository implements CartRepository backed by Redis.
// Carts are stored as JSON blobs under the key "cart:{userID}".
type RedisRepository struct {
	client redis.Cmdable
	ttl    time.Duration
	tracer trace.Tracer
}

// NewRedisRepository creates a Redis-backed repository with the given TTL.
// If ttl is zero or negative, defaultTTL (7 days) is used.
func NewRedisRepository(client redis.Cmdable, ttl time.Duration) *RedisRepository {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &RedisRepository{
		client: client,
		ttl:    ttl,
		tracer: otel.Tracer("cart/repository"),
	}
}

func cartKey(userID string) string {
	return keyPrefix + userID
}

// Get retrieves a cart from Redis.
func (r *RedisRepository) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	ctx, span := r.tracer.Start(ctx, "repository.Get")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	data, err := r.client.Get(ctx, cartKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
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

// Save serialises a cart to JSON and stores it in Redis with the configured TTL.
func (r *RedisRepository) Save(ctx context.Context, cart *domain.Cart) error {
	ctx, span := r.tracer.Start(ctx, "repository.Save")
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

// Delete removes the cart key from Redis.
func (r *RedisRepository) Delete(ctx context.Context, userID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.Delete")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	if err := r.client.Del(ctx, cartKey(userID)).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}
