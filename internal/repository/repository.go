// Package repository defines the storage interface for cart data.
package repository

import (
	"context"

	"github.com/marketplace/cart-service/internal/domain"
)

// CartRepository is the interface for cart storage operations.
type CartRepository interface {
	// Get retrieves a cart by user ID. Returns domain.ErrCartNotFound if absent.
	Get(ctx context.Context, userID string) (*domain.Cart, error)
	// Save persists a cart with the configured TTL, creating or overwriting.
	Save(ctx context.Context, cart *domain.Cart) error
	// Delete removes a cart by user ID. Does not error if the cart does not exist.
	Delete(ctx context.Context, userID string) error
}
