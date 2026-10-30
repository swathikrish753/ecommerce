package main

import (
	"context"

	"github.com/swathikrish753/ecommerce/pkg/config"
	"github.com/swathikrish753/ecommerce/pkg/httpserver"
	"github.com/swathikrish753/ecommerce/pkg/logger"
	pkgmongo "github.com/swathikrish753/ecommerce/pkg/mongo"
	"github.com/swathikrish753/ecommerce/services/catalog/internal/handler"
	"github.com/swathikrish753/ecommerce/services/catalog/internal/repository"
	"github.com/swathikrish753/ecommerce/services/catalog/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel, cfg.IsProd())
	log.WithField("env", cfg.Env).Info("starting catalog service")

	srv := httpserver.New(cfg.HTTPPort, log)

	ctx := context.Background()
	client, db, err := pkgmongo.New(ctx, pkgmongo.Config{
		URI:      cfg.MongoURI,
		Database: cfg.MongoDB,
	})
	if err != nil {
		log.WithError(err).Fatal("could not connect to mongo")
	}
	defer client.Disconnect(ctx)
	log.Info("connected to mongo")

	productRepo := repository.NewProductMongo(db)
	if err := productRepo.EnsureIndexes(ctx); err != nil {
		log.WithError(err).Fatal("could not create indexes")
	}

	catalogSvc := service.NewCatalog(productRepo)
	handler.NewCatalogHandler(catalogSvc).Register(srv.Echo)

	srv.SetReady(true)

	if err := srv.Run(); err != nil {
		log.WithError(err).Fatal("server exited with error")
	}
}
