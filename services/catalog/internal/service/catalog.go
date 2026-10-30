package service

import (
	"context"
	"errors"

	"github.com/swathikrish753/ecommerce/services/catalog/internal/domain"
)

var ErrInvalidProduct = errors.New("invalid product")

type Catalog struct {
	repo        domain.ProductRepository
	defaultSize int64
	maxSize     int64
}

func NewCatalog(repo domain.ProductRepository) *Catalog {
	return &Catalog{repo: repo, defaultSize: 20, maxSize: 100}
}

func (c *Catalog) Create(ctx context.Context, p *domain.Product) error {
	if p.SKU == "" || p.Name == "" || p.PriceCents < 0 {
		return ErrInvalidProduct
	}
	return c.repo.Create(ctx, p)
}

func (c *Catalog) Get(ctx context.Context, id string) (*domain.Product, error) {
	return c.repo.GetByID(ctx, id)
}

func (c *Catalog) List(ctx context.Context, category, afterID string, limit int64) ([]*domain.Product, error) {
	if limit <= 0 {
		limit = c.defaultSize
	}
	if limit > c.maxSize {
		limit = c.maxSize
	}
	return c.repo.List(ctx, category, afterID, limit)
}
