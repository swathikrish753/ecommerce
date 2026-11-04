package repository_test

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/swathikrish753/ecommerce/services/catalog/internal/domain"
	"github.com/swathikrish753/ecommerce/services/catalog/internal/repository"
)

func setupMongo(t *testing.T) *mongo.Database {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:7",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		t.Fatalf("start mongo: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })

	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "27017")
	uri := "mongodb://" + host + ":" + port.Port()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(ctx) })
	return client.Database("ecom_catalog")
}

func TestProductMongo_CreateGetListUnique(t *testing.T) {
	db := setupMongo(t)
	repo := repository.NewProductMongo(db)
	ctx := context.Background()

	if err := repo.EnsureIndexes(ctx); err != nil {
		t.Fatalf("indexes: %v", err)
	}

	p := &domain.Product{SKU: "A-1", Name: "Thing", Category: "c", PriceCents: 100}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected generated id")
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.SKU != "A-1" {
		t.Fatalf("got sku %q", got.SKU)
	}

	dup := &domain.Product{SKU: "A-1", Name: "Dup", PriceCents: 1}
	if err := repo.Create(ctx, dup); err != domain.ErrSKUTaken {
		t.Fatalf("want ErrSKUTaken, got %v", err)
	}

	items, err := repo.List(ctx, "", "", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) < 1 {
		t.Fatalf("want >=1 item, got %d", len(items))
	}
}
