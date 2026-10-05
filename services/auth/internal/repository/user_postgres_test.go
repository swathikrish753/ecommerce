package repository_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/swathikrish753/ecommerce/services/auth/internal/domain"
	"github.com/swathikrish753/ecommerce/services/auth/internal/repository"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("ecom_auth"),
		tcpostgres.WithUsername("ecom"),
		tcpostgres.WithPassword("ecom_pass"),
		tcpostgres.WithInitScripts(
			filepath.Join("..", "..", "migrations", "000001_create_users.up.sql"),
		),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("start container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestUserPostgres_CreateAndGet(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewUserPostgres(pool)
	ctx := context.Background()

	u := &domain.User{Email: "a@b.com", PasswordHash: "hash"}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == "" {
		t.Fatal("expected generated id")
	}

	got, err := repo.GetByEmail(ctx, "a@b.com")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Email != "a@b.com" {
		t.Fatalf("got email %q", got.Email)
	}

	dup := &domain.User{Email: "a@b.com", PasswordHash: "hash2"}
	if err := repo.Create(ctx, dup); err != domain.ErrEmailTaken {
		t.Fatalf("want ErrEmailTaken, got %v", err)
	}
}
