package service

import (
	"context"
	"errors"

	"github.com/swathikrish753/ecommerce/services/cart/internal/domain"
)

var (
	ErrInvalidItem = errors.New("invalid cart item")
	ErrNoUser      = errors.New("missing user id")
)

// maxQuantity caps a single line so a typo can't put 10^9 items in a cart.
const maxQuantity = 99

type Cart struct {
	repo domain.CartRepository
}

func NewCart(repo domain.CartRepository) *Cart {
	return &Cart{repo: repo}
}

func (s *Cart) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	if userID == "" {
		return nil, ErrNoUser
	}
	return s.repo.Get(ctx, userID)
}

// AddItem adds qty of sku. If the line would exceed maxQuantity the increment is rolled back.
func (s *Cart) AddItem(ctx context.Context, userID, sku string, qty int64) (int64, error) {
	if userID == "" {
		return 0, ErrNoUser
	}
	if sku == "" || qty <= 0 || qty > maxQuantity {
		return 0, ErrInvalidItem
	}
	n, err := s.repo.AddItem(ctx, userID, sku, qty)
	if err != nil {
		return 0, err
	}
	if n > maxQuantity {
		if _, rbErr := s.repo.AddItem(ctx, userID, sku, -qty); rbErr != nil {
			return 0, rbErr
		}
		return 0, ErrInvalidItem
	}
	return n, nil
}

func (s *Cart) SetQuantity(ctx context.Context, userID, sku string, qty int64) error {
	if userID == "" {
		return ErrNoUser
	}
	if sku == "" || qty < 0 || qty > maxQuantity {
		return ErrInvalidItem
	}
	if qty == 0 {
		return s.repo.RemoveItem(ctx, userID, sku)
	}
	return s.repo.SetQuantity(ctx, userID, sku, qty)
}

func (s *Cart) RemoveItem(ctx context.Context, userID, sku string) error {
	if userID == "" {
		return ErrNoUser
	}
	if sku == "" {
		return ErrInvalidItem
	}
	return s.repo.RemoveItem(ctx, userID, sku)
}

func (s *Cart) Clear(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrNoUser
	}
	return s.repo.Clear(ctx, userID)
}
