package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/marketplace/marketplace-bucket/internal/core/domain"
	"github.com/marketplace/marketplace-bucket/internal/core/port"
	"github.com/marketplace/marketplace-bucket/internal/platform/logger"
)

type CartUseCase struct {
	repo   port.CartRepository
	log    *slog.Logger
	tracer trace.Tracer
}

func NewCart(repo port.CartRepository, log *slog.Logger) *CartUseCase {
	return &CartUseCase{
		repo:   repo,
		log:    log,
		tracer: otel.Tracer("cart/usecase"),
	}
}

func (uc *CartUseCase) AddItem(ctx context.Context, userID string, item *domain.Item) (*domain.Cart, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.AddItem")
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

	cart, err := uc.getOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get or create cart: %w", err)
	}

	cart.AddItem(item)

	if err := uc.repo.Save(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save cart: %w", err)
	}

	logger.FromContext(ctx, uc.log).Debug("item added", "user_id", userID, "product_id", item.ProductID)
	return cart, nil
}

func (uc *CartUseCase) RemoveItem(ctx context.Context, userID string, productID string) (*domain.Cart, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.RemoveItem")
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

	cart, err := uc.repo.Get(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := cart.RemoveItem(productID); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save cart: %w", err)
	}

	return cart, nil
}

func (uc *CartUseCase) UpdateQuantity(ctx context.Context, userID string, productID string, quantity int) (*domain.Cart, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.UpdateQuantity")
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

	cart, err := uc.repo.Get(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := cart.UpdateQuantity(productID, quantity); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, cart); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("save cart: %w", err)
	}

	return cart, nil
}

func (uc *CartUseCase) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	_, span := uc.tracer.Start(ctx, "usecase.GetCart")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	if err := validateUserID(userID); err != nil {
		return nil, err
	}

	cart, err := uc.repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			return domain.NewCart(userID), nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return cart, nil
}

func (uc *CartUseCase) ClearCart(ctx context.Context, userID string) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.ClearCart")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID))

	if err := validateUserID(userID); err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete cart: %w", err)
	}

	return nil
}

func (uc *CartUseCase) getOrCreate(ctx context.Context, userID string) (*domain.Cart, error) {
	cart, err := uc.repo.Get(ctx, userID)
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
