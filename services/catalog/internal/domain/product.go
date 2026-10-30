package domain

import (
	"context"
	"errors"
	"time"
)

// Product is the catalog entity. bson tags map fields to MongoDB.
type Product struct {
	ID          string            `bson:"_id,omitempty"`
	SKU         string            `bson:"sku"`
	Name        string            `bson:"name"`
	Description string            `bson:"description"`
	Category    string            `bson:"category"`
	PriceCents  int64             `bson:"price_cents"`
	Attributes  map[string]string `bson:"attributes"`
	CreatedAt   time.Time         `bson:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at"`
}

var (
	ErrProductNotFound = errors.New("product not found")
	ErrSKUTaken        = errors.New("sku already exists")
)

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, category, afterID string, limit int64) ([]*Product, error)
}
