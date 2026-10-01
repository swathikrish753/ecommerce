package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// Server wraps an Echo instance plus the metadata it needs to run.
type Server struct {
	Echo *echo.Echo
	log  *logrus.Entry
	port int

	// ready flips to true once dependencies (DB, brokers) are up.
	// Readiness probes check this; liveness does not.
	ready atomic.Bool
}

// New builds a Server with sane default middleware and health routes.
func New(port int, log *logrus.Entry) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware runs on every request, in order.
	e.Use(middleware.Recover())   // turn panics into 500s instead of crashing
	e.Use(middleware.RequestID()) // attach/propagate an X-Request-ID
	e.Use(middleware.Gzip())      // compress responses

	s := &Server{Echo: e, log: log, port: port}

	// Liveness: is the process alive? Always 200 if we can answer.
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Readiness: are we ready to serve traffic (deps connected)?
	e.GET("/readyz", func(c echo.Context) error {
		if !s.ready.Load() {
			return c.JSON(http.StatusServiceUnavailable,
				map[string]string{"status": "not ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	return s
}

// SetReady marks the service ready (or not) for traffic.
func (s *Server) SetReady(ready bool) { s.ready.Store(ready) }

// Run starts the server and blocks until a shutdown signal arrives,
// then drains in-flight requests within a timeout.
func (s *Server) Run() error {
	go func() {
		addr := fmt.Sprintf(":%d", s.port)
		s.log.WithField("addr", addr).Info("http server starting")
		if err := s.Echo.Start(addr); err != nil && err != http.ErrServerClosed {
			s.log.WithError(err).Fatal("http server failed")
		}
	}()

	// Block until SIGINT (Ctrl-C) or SIGTERM (k8s pod termination).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.log.Info("shutdown signal received")

	// Give in-flight requests up to 10s to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	s.log.Info("http server stopped cleanly")
	return nil
}
