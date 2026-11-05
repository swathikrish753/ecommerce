package repository_test

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/swathikrish753/ecommerce/services/cart/internal/domain"
	"github.com/swathikrish753/ecommerce/services/cart/internal/repository"
)

func setupRedis(t *testing.T) *goredis.Client {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req, Started: true,
	})
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })

	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "6379")

	rdb := goredis.NewClient(&goredis.Options{Addr: host + ":" + port.Port()})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func TestCartRedis_AddGetSetRemoveClear(t *testing.T) {
	rdb := setupRedis(t)
	repo := repository.NewCartRedis(rdb, time.Hour)
	ctx := context.Background()
	user := "u-1"

	if n, err := repo.AddItem(ctx, user, "A-1", 2); err != nil || n != 2 {
		t.Fatalf("add: n=%d err=%v", n, err)
	}
	if n, err := repo.AddItem(ctx, user, "A-1", 3); err != nil || n != 5 {
		t.Fatalf("increment: n=%d err=%v", n, err)
	}
	if _, err := repo.AddItem(ctx, user, "B-2", 1); err != nil {
		t.Fatalf("add b: %v", err)
	}

	ttl, err := rdb.TTL(ctx, "cart:"+user).Result()
	if err != nil || ttl <= 0 {
		t.Fatalf("want ttl set, got %v err=%v", ttl, err)
	}

	cart, err := repo.Get(ctx, user)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(cart.Items) != 2 || cart.Items[0].SKU != "A-1" || cart.Items[0].Quantity != 5 {
		t.Fatalf("unexpected cart: %+v", cart.Items)
	}

	if err := repo.SetQuantity(ctx, user, "B-2", 7); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := repo.SetQuantity(ctx, user, "NOPE", 1); err != domain.ErrItemNotFound {
		t.Fatalf("want ErrItemNotFound, got %v", err)
	}

	if err := repo.RemoveItem(ctx, user, "A-1"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := repo.RemoveItem(ctx, user, "A-1"); err != domain.ErrItemNotFound {
		t.Fatalf("want ErrItemNotFound on second remove, got %v", err)
	}

	if err := repo.Clear(ctx, user); err != nil {
		t.Fatalf("clear: %v", err)
	}
	cart, err = repo.Get(ctx, user)
	if err != nil || len(cart.Items) != 0 {
		t.Fatalf("want empty cart, got %+v err=%v", cart, err)
	}
}

func TestCartRedis_ConcurrentAddIsAtomic(t *testing.T) {
	rdb := setupRedis(t)
	repo := repository.NewCartRedis(rdb, time.Hour)
	ctx := context.Background()

	const workers = 50
	done := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			_, err := repo.AddItem(ctx, "u-race", "SKU", 1)
			done <- err
		}()
	}
	for i := 0; i < workers; i++ {
		if err := <-done; err != nil {
			t.Fatalf("add: %v", err)
		}
	}
	cart, err := repo.Get(ctx, "u-race")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(cart.Items) != 1 || cart.Items[0].Quantity != workers {
		t.Fatalf("want quantity %d, got %+v", workers, cart.Items)
	}
}
