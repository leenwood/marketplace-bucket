package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/marketplace/cart-service/internal/domain"
	"github.com/marketplace/cart-service/internal/service"
	"github.com/marketplace/cart-service/pkg/metrics"
)

// mockRepo is a testify mock for repository.CartRepository.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cart), args.Error(1)
}

func (m *mockRepo) Save(ctx context.Context, cart *domain.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *mockRepo) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func newTestService(r *mockRepo) service.CartService {
	return service.NewCartService(r, metrics.New("test"))
}

// ── AddItem ───────────────────────────────────────────────────────────────────

func TestCartService_AddItem(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		item      *domain.Item
		setupMock func(*mockRepo)
		wantErr   error
	}{
		{
			name:   "new item on empty cart",
			userID: "u1",
			item:   &domain.Item{ProductID: "p1", Name: "Widget", Price: 9.99, Quantity: 2},
			setupMock: func(r *mockRepo) {
				r.On("Get", mock.Anything, "u1").Return(nil, domain.ErrCartNotFound)
				r.On("Save", mock.Anything, mock.AnythingOfType("*domain.Cart")).Return(nil)
			},
		},
		{
			name:   "add to existing cart",
			userID: "u1",
			item:   &domain.Item{ProductID: "p2", Name: "Gadget", Price: 19.99, Quantity: 1},
			setupMock: func(r *mockRepo) {
				c := domain.NewCart("u1")
				c.Items["p1"] = &domain.Item{ProductID: "p1", Quantity: 3}
				r.On("Get", mock.Anything, "u1").Return(c, nil)
				r.On("Save", mock.Anything, mock.AnythingOfType("*domain.Cart")).Return(nil)
			},
		},
		{
			name:      "empty user ID",
			userID:    "",
			item:      &domain.Item{ProductID: "p1", Quantity: 1},
			setupMock: func(_ *mockRepo) {},
			wantErr:   domain.ErrInvalidUserID,
		},
		{
			name:      "empty product ID",
			userID:    "u1",
			item:      &domain.Item{ProductID: "", Quantity: 1},
			setupMock: func(_ *mockRepo) {},
			wantErr:   domain.ErrInvalidProductID,
		},
		{
			name:      "zero quantity",
			userID:    "u1",
			item:      &domain.Item{ProductID: "p1", Quantity: 0},
			setupMock: func(_ *mockRepo) {},
			wantErr:   domain.ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockRepo{}
			tt.setupMock(r)
			svc := newTestService(r)

			cart, err := svc.AddItem(context.Background(), tt.userID, tt.item)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cart)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cart)
			}
			r.AssertExpectations(t)
		})
	}
}

// ── RemoveItem ────────────────────────────────────────────────────────────────

func TestCartService_RemoveItem(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		productID string
		setupMock func(*mockRepo)
		wantErr   error
	}{
		{
			name:      "remove existing item",
			userID:    "u1",
			productID: "p1",
			setupMock: func(r *mockRepo) {
				c := domain.NewCart("u1")
				c.Items["p1"] = &domain.Item{ProductID: "p1", Quantity: 2}
				r.On("Get", mock.Anything, "u1").Return(c, nil)
				r.On("Save", mock.Anything, mock.AnythingOfType("*domain.Cart")).Return(nil)
			},
		},
		{
			name:      "item not in cart",
			userID:    "u1",
			productID: "missing",
			setupMock: func(r *mockRepo) {
				r.On("Get", mock.Anything, "u1").Return(domain.NewCart("u1"), nil)
			},
			wantErr: domain.ErrItemNotFound,
		},
		{
			name:      "empty user ID",
			userID:    "",
			productID: "p1",
			setupMock: func(_ *mockRepo) {},
			wantErr:   domain.ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockRepo{}
			tt.setupMock(r)
			svc := newTestService(r)

			cart, err := svc.RemoveItem(context.Background(), tt.userID, tt.productID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cart)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cart)
			}
			r.AssertExpectations(t)
		})
	}
}

// ── UpdateQuantity ────────────────────────────────────────────────────────────

func TestCartService_UpdateQuantity(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		productID string
		quantity  int
		setupMock func(*mockRepo)
		wantErr   error
	}{
		{
			name:      "valid update",
			userID:    "u1",
			productID: "p1",
			quantity:  5,
			setupMock: func(r *mockRepo) {
				c := domain.NewCart("u1")
				c.Items["p1"] = &domain.Item{ProductID: "p1", Quantity: 1}
				r.On("Get", mock.Anything, "u1").Return(c, nil)
				r.On("Save", mock.Anything, mock.AnythingOfType("*domain.Cart")).Return(nil)
			},
		},
		{
			name:      "zero quantity rejected",
			userID:    "u1",
			productID: "p1",
			quantity:  0,
			setupMock: func(_ *mockRepo) {},
			wantErr:   domain.ErrInvalidQuantity,
		},
		{
			name:      "item not found",
			userID:    "u1",
			productID: "missing",
			quantity:  3,
			setupMock: func(r *mockRepo) {
				r.On("Get", mock.Anything, "u1").Return(domain.NewCart("u1"), nil)
			},
			wantErr: domain.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockRepo{}
			tt.setupMock(r)
			svc := newTestService(r)

			cart, err := svc.UpdateQuantity(context.Background(), tt.userID, tt.productID, tt.quantity)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cart)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.quantity, cart.Items[tt.productID].Quantity)
			}
			r.AssertExpectations(t)
		})
	}
}

// ── GetCart ───────────────────────────────────────────────────────────────────

func TestCartService_GetCart(t *testing.T) {
	t.Run("returns existing cart", func(t *testing.T) {
		r := &mockRepo{}
		c := domain.NewCart("u1")
		c.Items["p1"] = &domain.Item{ProductID: "p1", Quantity: 1}
		r.On("Get", mock.Anything, "u1").Return(c, nil)

		cart, err := newTestService(r).GetCart(context.Background(), "u1")
		require.NoError(t, err)
		assert.Len(t, cart.Items, 1)
		r.AssertExpectations(t)
	})

	t.Run("returns empty cart when not found", func(t *testing.T) {
		r := &mockRepo{}
		r.On("Get", mock.Anything, "u1").Return(nil, domain.ErrCartNotFound)

		cart, err := newTestService(r).GetCart(context.Background(), "u1")
		require.NoError(t, err)
		assert.Empty(t, cart.Items)
		r.AssertExpectations(t)
	})
}

// ── ClearCart ─────────────────────────────────────────────────────────────────

func TestCartService_ClearCart(t *testing.T) {
	t.Run("deletes cart", func(t *testing.T) {
		r := &mockRepo{}
		r.On("Delete", mock.Anything, "u1").Return(nil)

		require.NoError(t, newTestService(r).ClearCart(context.Background(), "u1"))
		r.AssertExpectations(t)
	})

	t.Run("empty user ID", func(t *testing.T) {
		r := &mockRepo{}
		err := newTestService(r).ClearCart(context.Background(), "")
		require.ErrorIs(t, err, domain.ErrInvalidUserID)
	})
}
