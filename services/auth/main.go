package main

import (
	"context"
	"time"

	"github.com/swathikrish753/ecommerce/pkg/config"
	"github.com/swathikrish753/ecommerce/pkg/httpserver"
	"github.com/swathikrish753/ecommerce/pkg/logger"
	"github.com/swathikrish753/ecommerce/pkg/postgres"
	"github.com/swathikrish753/ecommerce/services/auth/internal/handler"
	"github.com/swathikrish753/ecommerce/services/auth/internal/repository"
	"github.com/swathikrish753/ecommerce/services/auth/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel, cfg.IsProd())
	log.WithField("env", cfg.Env).Info("starting auth service")

	srv := httpserver.New(cfg.HTTPPort, log)

	ctx := context.Background()
	pool, err := postgres.New(ctx, postgres.Config{
		DSN:             cfg.DBDSN,
		MaxConns:        10,
		MaxConnLifetime: time.Hour,
	})
	if err != nil {
		log.WithError(err).Fatal("could not connect to postgres")
	}
	defer pool.Close()
	log.Info("connected to postgres")

	userRepo := repository.NewUserPostgres(pool)
	authSvc := service.NewAuth(userRepo, cfg.JWTSecret,
		time.Duration(cfg.JWTTTLMinutes)*time.Minute)
	authHandler := handler.NewAuthHandler(authSvc)
	authHandler.Register(srv.Echo)

	srv.SetReady(true)

	if err := srv.Run(); err != nil {
		log.WithError(err).Fatal("server exited with error")
	}
}
