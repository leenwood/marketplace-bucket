package pb

// Item is the gRPC representation of a cart item.
type Item struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int32   `json:"quantity"`
}

// CartResponse is the standard cart payload returned by all mutating RPCs.
type CartResponse struct {
	UserID string  `json:"user_id"`
	Items  []*Item `json:"items"`
	Total  float64 `json:"total"`
}

// AddItemRequest is the request body for AddItem.
type AddItemRequest struct {
	UserID string `json:"user_id"`
	Item   *Item  `json:"item"`
}

// RemoveItemRequest is the request body for RemoveItem.
type RemoveItemRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
}

// UpdateQuantityRequest is the request body for UpdateQuantity.
type UpdateQuantityRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

// GetCartRequest is the request body for GetCart.
type GetCartRequest struct {
	UserID string `json:"user_id"`
}

// ClearCartRequest is the request body for ClearCart.
type ClearCartRequest struct {
	UserID string `json:"user_id"`
}
