package main

import (
	"github.com/swathikrish753/ecommerce/pkg/config"
	"github.com/swathikrish753/ecommerce/pkg/httpserver"
	"github.com/swathikrish753/ecommerce/pkg/logger"
)

func main() {
	// 1. Load configuration from the environment (Viper).
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// 2. Build a structured logger.
	log := logger.New(cfg.ServiceName, cfg.LogLevel, cfg.IsProd())
	log.WithField("env", cfg.Env).Info("starting auth service")

	// 3. Build the HTTP server.
	srv := httpserver.New(cfg.HTTPPort, log)

	// 4. Mark ready. (From Day 3 we flip this true only AFTER the DB
	//    connection succeeds — for now there are no deps, so ready now.)
	srv.SetReady(true)

	// 5. Run until a shutdown signal, then drain gracefully.
	if err := srv.Run(); err != nil {
		log.WithError(err).Fatal("server exited with error")
	}
}
