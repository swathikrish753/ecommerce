package main

import (
	"context"
	"time"

	"github.com/swathikrish753/ecommerce/pkg/config"
	"github.com/swathikrish753/ecommerce/pkg/httpserver"
	"github.com/swathikrish753/ecommerce/pkg/logger"
	pkgredis "github.com/swathikrish753/ecommerce/pkg/redis"
	"github.com/swathikrish753/ecommerce/services/cart/internal/handler"
	"github.com/swathikrish753/ecommerce/services/cart/internal/repository"
	"github.com/swathikrish753/ecommerce/services/cart/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel, cfg.IsProd())
	log.WithField("env", cfg.Env).Info("starting cart service")

	srv := httpserver.New(cfg.HTTPPort, log)

	ctx := context.Background()
	rdb, err := pkgredis.New(ctx, pkgredis.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		log.WithError(err).Fatal("could not connect to redis")
	}
	defer rdb.Close()
	log.Info("connected to redis")

	ttl := time.Duration(cfg.CartTTLHours) * time.Hour
	cartRepo := repository.NewCartRedis(rdb, ttl)
	cartSvc := service.NewCart(cartRepo)
	handler.NewCartHandler(cartSvc).Register(srv.Echo)

	srv.SetReady(true)

	if err := srv.Run(); err != nil {
		log.WithError(err).Fatal("server exited with error")
	}
}
