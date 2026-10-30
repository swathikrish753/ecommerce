package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds MongoDB connection settings.
type Config struct {
	URI      string
	Database string
}

// New connects to MongoDB and verifies it with a ping.
// Caller is responsible for client.Disconnect(ctx) on shutdown.
func New(ctx context.Context, cfg Config) (*mongo.Client, *mongo.Database, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(connectCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("ping mongo: %w", err)
	}

	return client, client.Database(cfg.Database), nil
}
