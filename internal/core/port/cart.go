package port

import (
	"context"

	"github.com/marketplace/marketplace-bucket/internal/core/domain"
)

type CartRepository interface {
	Get(ctx context.Context, userID string) (*domain.Cart, error)
	Save(ctx context.Context, cart *domain.Cart) error
	Delete(ctx context.Context, userID string) error
}

type CartService interface {
	AddItem(ctx context.Context, userID string, item *domain.Item) (*domain.Cart, error)
	RemoveItem(ctx context.Context, userID string, productID string) (*domain.Cart, error)
	UpdateQuantity(ctx context.Context, userID string, productID string, quantity int) (*domain.Cart, error)
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	ClearCart(ctx context.Context, userID string) error
	MergeCart(ctx context.Context, guestUserID, targetUserID string) (*domain.Cart, error)
}
