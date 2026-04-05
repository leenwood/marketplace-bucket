package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marketplace/cart-service/internal/domain"
	"github.com/marketplace/cart-service/internal/repository"
)

func setupMiniRedis(t *testing.T) (*repository.RedisRepository, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return repository.NewRedisRepository(client, 24*time.Hour), mr
}

func TestRedisRepository_SaveAndGet(t *testing.T) {
	repo, _ := setupMiniRedis(t)
	ctx := context.Background()

	cart := domain.NewCart("user1")
	cart.Items["prod1"] = &domain.Item{
		ProductID: "prod1",
		Name:      "Widget",
		Price:     9.99,
		Quantity:  2,
	}

	require.NoError(t, repo.Save(ctx, cart))

	got, err := repo.Get(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, cart.UserID, got.UserID)
	assert.Len(t, got.Items, 1)
	assert.Equal(t, cart.Items["prod1"].Name, got.Items["prod1"].Name)
	assert.InDelta(t, cart.Items["prod1"].Price, got.Items["prod1"].Price, 0.001)
}

func TestRedisRepository_GetNotFound(t *testing.T) {
	repo, _ := setupMiniRedis(t)

	_, err := repo.Get(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrCartNotFound)
}

func TestRedisRepository_Delete(t *testing.T) {
	repo, _ := setupMiniRedis(t)
	ctx := context.Background()

	cart := domain.NewCart("user1")
	require.NoError(t, repo.Save(ctx, cart))
	require.NoError(t, repo.Delete(ctx, "user1"))

	_, err := repo.Get(ctx, "user1")
	assert.ErrorIs(t, err, domain.ErrCartNotFound)
}

func TestRedisRepository_TTL(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	ttl := 5 * time.Second
	repo := repository.NewRedisRepository(client, ttl)
	ctx := context.Background()

	cart := domain.NewCart("user_ttl")
	require.NoError(t, repo.Save(ctx, cart))

	mr.FastForward(ttl + time.Second)

	_, err = repo.Get(ctx, "user_ttl")
	assert.ErrorIs(t, err, domain.ErrCartNotFound)
}

func TestRedisRepository_OverwriteCart(t *testing.T) {
	repo, _ := setupMiniRedis(t)
	ctx := context.Background()

	cart := domain.NewCart("user1")
	cart.Items["prod1"] = &domain.Item{ProductID: "prod1", Quantity: 1}
	require.NoError(t, repo.Save(ctx, cart))

	cart.Items["prod2"] = &domain.Item{ProductID: "prod2", Quantity: 3}
	require.NoError(t, repo.Save(ctx, cart))

	got, err := repo.Get(ctx, "user1")
	require.NoError(t, err)
	assert.Len(t, got.Items, 2)
}
