package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/web3-frozen/demo-api/internal/cache"
	"github.com/web3-frozen/demo-api/internal/handler"
	"github.com/web3-frozen/demo-api/internal/middleware"
	"github.com/web3-frozen/demo-api/internal/queue"
	"github.com/web3-frozen/demo-api/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	port := envOr("PORT", "8080")
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := envOr("KAFKA_TOPIC", "task-events")

	if dbURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	// PostgreSQL
	pg, err := store.NewPostgresStore(dbURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	if err := pg.Migrate(context.Background()); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected and migrated")

	// Redis (optional — degrades gracefully)
	var rc *cache.RedisCache
	if redisURL != "" {
		rc, err = cache.NewRedisCache(redisURL)
		if err != nil {
			logger.Warn("redis unavailable, caching disabled", "error", err)
		} else {
			defer rc.Close()
			logger.Info("redis connected")
		}
	}

	// Kafka (optional — degrades gracefully)
	var kp *queue.KafkaProducer
	if kafkaBrokers != "" {
		kp = queue.NewKafkaProducer(kafkaBrokers, kafkaTopic, logger)
		defer kp.Close()
		logger.Info("kafka producer initialized", "brokers", kafkaBrokers, "topic", kafkaTopic)
	}

	// Routes
	taskHandler := handler.NewTaskHandler(pg, rc, kp, logger)
	r := chi.NewRouter()
	r.Use(middleware.Recover(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pg.Ping(r.Context()); err != nil {
			http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})
	r.Mount("/api/tasks", taskHandler.Routes())

	// Server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
