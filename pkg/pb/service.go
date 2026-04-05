package pb

import (
	"context"

	"google.golang.org/grpc"
)

// CartServiceServer is the server-side interface for the cart gRPC service.
type CartServiceServer interface {
	AddItem(ctx context.Context, req *AddItemRequest) (*CartResponse, error)
	RemoveItem(ctx context.Context, req *RemoveItemRequest) (*CartResponse, error)
	UpdateQuantity(ctx context.Context, req *UpdateQuantityRequest) (*CartResponse, error)
	GetCart(ctx context.Context, req *GetCartRequest) (*CartResponse, error)
	ClearCart(ctx context.Context, req *ClearCartRequest) (*CartResponse, error)
}

// RegisterCartServiceServer attaches srv to the gRPC server s.
func RegisterCartServiceServer(s *grpc.Server, srv CartServiceServer) {
	s.RegisterService(&cartServiceDesc, srv)
}

var cartServiceDesc = grpc.ServiceDesc{
	ServiceName: "cart.v1.CartService",
	HandlerType: (*CartServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "AddItem", Handler: _CartService_AddItem_Handler},
		{MethodName: "RemoveItem", Handler: _CartService_RemoveItem_Handler},
		{MethodName: "UpdateQuantity", Handler: _CartService_UpdateQuantity_Handler},
		{MethodName: "GetCart", Handler: _CartService_GetCart_Handler},
		{MethodName: "ClearCart", Handler: _CartService_ClearCart_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cart.proto",
}

func _CartService_AddItem_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	req := new(AddItemRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartServiceServer).AddItem(ctx, req)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/cart.v1.CartService/AddItem"}
	return interceptor(ctx, req, info, func(ctx context.Context, r any) (any, error) {
		return srv.(CartServiceServer).AddItem(ctx, r.(*AddItemRequest))
	})
}

func _CartService_RemoveItem_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	req := new(RemoveItemRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartServiceServer).RemoveItem(ctx, req)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/cart.v1.CartService/RemoveItem"}
	return interceptor(ctx, req, info, func(ctx context.Context, r any) (any, error) {
		return srv.(CartServiceServer).RemoveItem(ctx, r.(*RemoveItemRequest))
	})
}

func _CartService_UpdateQuantity_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	req := new(UpdateQuantityRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartServiceServer).UpdateQuantity(ctx, req)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/cart.v1.CartService/UpdateQuantity"}
	return interceptor(ctx, req, info, func(ctx context.Context, r any) (any, error) {
		return srv.(CartServiceServer).UpdateQuantity(ctx, r.(*UpdateQuantityRequest))
	})
}

func _CartService_GetCart_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	req := new(GetCartRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartServiceServer).GetCart(ctx, req)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/cart.v1.CartService/GetCart"}
	return interceptor(ctx, req, info, func(ctx context.Context, r any) (any, error) {
		return srv.(CartServiceServer).GetCart(ctx, r.(*GetCartRequest))
	})
}

func _CartService_ClearCart_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	req := new(ClearCartRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartServiceServer).ClearCart(ctx, req)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/cart.v1.CartService/ClearCart"}
	return interceptor(ctx, req, info, func(ctx context.Context, r any) (any, error) {
		return srv.(CartServiceServer).ClearCart(ctx, r.(*ClearCartRequest))
	})
}
