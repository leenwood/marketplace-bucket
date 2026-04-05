// Package domain contains core business types for the cart service.
package domain

import (
	"errors"
	"time"
)

// Sentinel errors for cart operations.
var (
	ErrCartNotFound     = errors.New("cart not found")
	ErrItemNotFound     = errors.New("item not found in cart")
	ErrInvalidQuantity  = errors.New("quantity must be greater than zero")
	ErrInvalidUserID    = errors.New("user ID cannot be empty")
	ErrInvalidProductID = errors.New("product ID cannot be empty")
)

// Item represents a product in the shopping cart.
type Item struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// Cart represents a user's shopping cart.
type Cart struct {
	UserID    string           `json:"user_id"`
	Items     map[string]*Item `json:"items"` // keyed by ProductID
	UpdatedAt time.Time        `json:"updated_at"`
}

// NewCart creates a new empty cart for the given user.
func NewCart(userID string) *Cart {
	return &Cart{
		UserID:    userID,
		Items:     make(map[string]*Item),
		UpdatedAt: time.Now(),
	}
}

// Total returns the sum of price * quantity for all items.
func (c *Cart) Total() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

// AddItem adds an item to the cart, accumulating quantity if it already exists.
func (c *Cart) AddItem(item *Item) {
	if existing, ok := c.Items[item.ProductID]; ok {
		existing.Quantity += item.Quantity
		existing.Name = item.Name
		existing.Price = item.Price
	} else {
		copied := *item
		c.Items[item.ProductID] = &copied
	}
	c.UpdatedAt = time.Now()
}

// RemoveItem removes an item by product ID. Returns ErrItemNotFound if absent.
func (c *Cart) RemoveItem(productID string) error {
	if _, ok := c.Items[productID]; !ok {
		return ErrItemNotFound
	}
	delete(c.Items, productID)
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateQuantity sets the quantity of an existing item. Returns ErrItemNotFound if absent.
func (c *Cart) UpdateQuantity(productID string, quantity int) error {
	item, ok := c.Items[productID]
	if !ok {
		return ErrItemNotFound
	}
	item.Quantity = quantity
	c.UpdatedAt = time.Now()
	return nil
}
