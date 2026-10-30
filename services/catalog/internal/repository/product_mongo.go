package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/swathikrish753/ecommerce/services/catalog/internal/domain"
)

type ProductMongo struct {
	coll *mongo.Collection
}

func NewProductMongo(db *mongo.Database) *ProductMongo {
	return &ProductMongo{coll: db.Collection("products")}
}

func (r *ProductMongo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sku", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create sku index: %w", err)
	}
	return nil
}

func (r *ProductMongo) Create(ctx context.Context, p *domain.Product) error {
	now := time.Now().UTC()
	p.CreatedAt, p.UpdatedAt = now, now

	doc := bson.M{
		"sku":         p.SKU,
		"name":        p.Name,
		"description": p.Description,
		"category":    p.Category,
		"price_cents": p.PriceCents,
		"attributes":  p.Attributes,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}

	res, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrSKUTaken
		}
		return fmt.Errorf("insert product: %w", err)
	}
	p.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *ProductMongo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrProductNotFound
	}

	var p domain.Product
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("find product: %w", err)
	}
	p.ID = oid.Hex()
	return &p, nil
}

func (r *ProductMongo) List(ctx context.Context, category, afterID string, limit int64) ([]*domain.Product, error) {
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}
	if afterID != "" {
		oid, err := primitive.ObjectIDFromHex(afterID)
		if err == nil {
			filter["_id"] = bson.M{"$gt": oid}
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(limit)

	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer cur.Close(ctx)

	var out []*domain.Product
	for cur.Next(ctx) {
		var p domain.Product
		if err := cur.Decode(&p); err != nil {
			return nil, fmt.Errorf("decode product: %w", err)
		}
		out = append(out, &p)
	}
	return out, cur.Err()
}
