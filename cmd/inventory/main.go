package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/web3-frozen/demo-api/internal/cache"
	"github.com/web3-frozen/demo-api/internal/downstream"
	"github.com/web3-frozen/demo-api/internal/inventory"
	"github.com/web3-frozen/demo-api/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	port := envOr("PORT", "8081")

	pg, err := store.NewPostgresStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pg.Close()
	if err := pg.Migrate(context.Background()); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	var reservationCache *cache.CheckoutCache
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		reservationCache, err = cache.NewCheckoutCache(redisURL)
		if err != nil {
			logger.Warn("redis unavailable; reservation cache disabled", "error", err)
		} else {
			defer func() { _ = reservationCache.Close() }()
		}
	}

	pricingClient, err := downstream.NewClient(envOr("PRICING_URL", "http://demo-pricing.demo-app.svc.cluster.local:8082"))
	if err != nil {
		logger.Error("invalid pricing service URL", "error", err)
		os.Exit(1)
	}

	service := inventory.NewService(pg, reservationCache, pricingClient, logger)
	serve(port, service.Handler(), logger)
}

func serve(port string, handler http.Handler, logger *slog.Logger) {
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("inventory service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("inventory service failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
