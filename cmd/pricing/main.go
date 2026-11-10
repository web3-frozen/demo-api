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
	"github.com/web3-frozen/demo-api/internal/pricing"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	port := envOr("PORT", "8082")

	var quoteCache *cache.CheckoutCache
	var err error
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		quoteCache, err = cache.NewCheckoutCache(redisURL)
		if err != nil {
			logger.Warn("redis unavailable; quote cache disabled", "error", err)
		} else {
			defer func() { _ = quoteCache.Close() }()
		}
	}

	providerURL := envOr("MARKET_DATA_URL", "https://api.frankfurter.app/latest?from=USD&to=EUR")
	service := pricing.NewService(providerURL, quoteCache, logger)
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
		logger.Info("pricing service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("pricing service failed", "error", err)
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
