// Package grpc provides the gRPC transport layer for the cart service.
package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/marketplace/marketplace-bucket/internal/domain"
	"github.com/marketplace/marketplace-bucket/internal/service"
	"github.com/marketplace/marketplace-bucket/pkg/metrics"
	"github.com/marketplace/marketplace-bucket/pkg/pb"
)

// Handler implements pb.CartServiceServer.
type Handler struct {
	svc     service.CartService
	metrics *metrics.Metrics
}

// NewHandler creates a new gRPC cart handler.
func NewHandler(svc service.CartService, m *metrics.Metrics) *Handler {
	return &Handler{svc: svc, metrics: m}
}

// AddItem handles the AddItem RPC.
func (h *Handler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.CartResponse, error) {
	if req.Item == nil {
		return nil, status.Error(codes.InvalidArgument, "item is required")
	}

	item := &domain.Item{
		ProductID: req.Item.ProductID,
		Name:      req.Item.Name,
		Price:     req.Item.Price,
		Quantity:  int(req.Item.Quantity),
	}

	cart, err := h.svc.AddItem(ctx, req.UserID, item)
	if err != nil {
		return nil, domainToGRPC(err)
	}
	return cartToProto(cart), nil
}

// RemoveItem handles the RemoveItem RPC.
func (h *Handler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.CartResponse, error) {
	cart, err := h.svc.RemoveItem(ctx, req.UserID, req.ProductID)
	if err != nil {
		return nil, domainToGRPC(err)
	}
	return cartToProto(cart), nil
}

// UpdateQuantity handles the UpdateQuantity RPC.
func (h *Handler) UpdateQuantity(ctx context.Context, req *pb.UpdateQuantityRequest) (*pb.CartResponse, error) {
	cart, err := h.svc.UpdateQuantity(ctx, req.UserID, req.ProductID, int(req.Quantity))
	if err != nil {
		return nil, domainToGRPC(err)
	}
	return cartToProto(cart), nil
}

// GetCart handles the GetCart RPC.
func (h *Handler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
	cart, err := h.svc.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, domainToGRPC(err)
	}
	return cartToProto(cart), nil
}

// ClearCart handles the ClearCart RPC.
func (h *Handler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.CartResponse, error) {
	if err := h.svc.ClearCart(ctx, req.UserID); err != nil {
		return nil, domainToGRPC(err)
	}
	return &pb.CartResponse{UserID: req.UserID, Items: []*pb.Item{}}, nil
}

// cartToProto converts a domain cart to its protobuf representation.
func cartToProto(cart *domain.Cart) *pb.CartResponse {
	items := make([]*pb.Item, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, &pb.Item{
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  int32(item.Quantity),
		})
	}
	return &pb.CartResponse{
		UserID: cart.UserID,
		Items:  items,
		Total:  cart.Total(),
	}
}

// domainToGRPC maps domain sentinel errors to gRPC status codes.
func domainToGRPC(err error) error {
	switch {
	case errors.Is(err, domain.ErrCartNotFound), errors.Is(err, domain.ErrItemNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidProductID):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
