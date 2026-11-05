package domain

import (
	"context"
	"errors"
)

// CartItem is one line in a user's cart: a SKU and how many of it.
type CartItem struct {
	SKU      string `json:"sku"`
	Quantity int64  `json:"quantity"`
}

// Cart is the full cart for one user.
type Cart struct {
	UserID string      `json:"user_id"`
	Items  []*CartItem `json:"items"`
}

var ErrItemNotFound = errors.New("item not in cart")

type CartRepository interface {
	Get(ctx context.Context, userID string) (*Cart, error)
	// AddItem atomically increments the quantity of sku by delta and returns the new quantity.
	AddItem(ctx context.Context, userID, sku string, delta int64) (int64, error)
	// SetQuantity overwrites the quantity of an existing sku.
	SetQuantity(ctx context.Context, userID, sku string, qty int64) error
	RemoveItem(ctx context.Context, userID, sku string) error
	Clear(ctx context.Context, userID string) error
}
