// Package service implements cart business logic.
package service

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/marketplace/marketplace-bucket/internal/domain"
	"github.com/marketplace/marketplace-bucket/internal/repository"
	"github.com/marketplace/marketplace-bucket/pkg/metrics"
)

// CartService is the interface for cart business operations.
type CartService interface {
	AddItem(ctx context.Context, userID string, item *domain.Item) (*domain.Cart, error)
	RemoveItem(ctx context.Context, userID string, productID string) (*domain.Cart, error)
	UpdateQuantity(ctx context.Context, userID string, productID string, quantity int) (*domain.Cart, error)
	GetCart(ctx context.Context, userID string) (*domain.Cart, error)
	ClearCart(ctx context.Context, userID string) error
}

type cartService struct {
	repo    repository.CartRepository
	metrics *metrics.Metrics
	tracer  trace.Tracer
}

// NewCartService constructs a CartService with the provided repository and metrics.
func NewCartService(repo repository.CartRepository, m *metrics.Metrics) CartService {
	return &cartService{
		repo:    repo,
		metrics: m,
		tracer:  otel.Tracer("cart/service"),
	}
}

// AddItem adds the item to the user's cart, creating the cart if it does not exist.
// If the product already exists, its quantity is incremented.
func (s *cartService) AddItem(ctx context.Context, userID string, item *domain.Item) (*domain.Cart, error) {
	ctx, span := s.tracer.Start(ctx, "service.AddItem")
	defer span.End()
	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("product_id", item.ProductID),
	)

	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	if err := validateItem(item); err != nil {
		return nil, err
	}

	cart, err := s.getOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get or create cart: %w", err)
	}

	cart.AddItem(item)

	if err := s.repo.Save(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save cart: %w", err)
	}

	s.metrics.CartOperationsTotal.WithLabelValues("add_item").Inc()
	return cart, nil
}

// RemoveItem removes the given product from the user's cart.
func (s *cartService) RemoveItem(ctx context.Context, userID string, productID string) (*domain.Cart, error) {
	ctx, span := s.tracer.Start(ctx, "service.RemoveItem")
	defer span.End()
	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("product_id", productID),
	)

	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	if productID == "" {
		return nil, domain.ErrInvalidProductID
	}

	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := cart.RemoveItem(productID); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save cart: %w", err)
	}

	s.metrics.CartOperationsTotal.WithLabelValues("remove_item").Inc()
	return cart, nil
}

// UpdateQuantity sets the quantity of an item in the user's cart.
func (s *cartService) UpdateQuantity(ctx context.Context, userID string, productID string, quantity int) (*domain.Cart, error) {
	ctx, span := s.tracer.Start(ctx, "service.UpdateQuantity")
	defer span.End()
	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("product_id", productID),
		attribute.Int("quantity", quantity),
	)

	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	if productID == "" {
		return nil, domain.ErrInvalidProductID
	}
	if quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := cart.UpdateQuantity(productID, quantity); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save cart: %w", err)
	}

	s.metrics.CartOperationsTotal.WithLabelValues("update_quantity").Inc()
	return cart, nil
}

// GetCart returns the user's cart, or an empty cart if none exists yet.
func (s *cartService) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	_, span := s.tracer.Start(ctx, "service.GetCart")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	if err := validateUserID(userID); err != nil {
		return nil, err
	}

	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			return domain.NewCart(userID), nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	s.metrics.CartOperationsTotal.WithLabelValues("get_cart").Inc()
	return cart, nil
}

// ClearCart deletes the user's cart.
func (s *cartService) ClearCart(ctx context.Context, userID string) error {
	ctx, span := s.tracer.Start(ctx, "service.ClearCart")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	if err := validateUserID(userID); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete cart: %w", err)
	}

	s.metrics.CartOperationsTotal.WithLabelValues("clear_cart").Inc()
	return nil
}

func (s *cartService) getOrCreate(ctx context.Context, userID string) (*domain.Cart, error) {
	cart, err := s.repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			return domain.NewCart(userID), nil
		}
		return nil, err
	}
	return cart, nil
}

func validateUserID(userID string) error {
	if userID == "" {
		return domain.ErrInvalidUserID
	}
	return nil
}

func validateItem(item *domain.Item) error {
	if item.ProductID == "" {
		return domain.ErrInvalidProductID
	}
	if item.Quantity <= 0 {
		return domain.ErrInvalidQuantity
	}
	return nil
}
