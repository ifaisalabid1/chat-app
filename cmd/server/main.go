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
	"github.com/ifaisalabid1/chat-app/internal/cache"
	"github.com/ifaisalabid1/chat-app/internal/config"
	"github.com/ifaisalabid1/chat-app/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.NewPool(ctx, cfg.DB)
	if err != nil {
		logger.Error("failed to create pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Error("redis client error: %w", err)
	}
	defer rdb.Close()

	userRepo := repository.NewUserRepository(pool)
	msgRepo := repository.NewMessageRepository(pool)

	r := chi.NewRouter()

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	go func() {
		logger.Info("server starting", "port", cfg.HTTP.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)
}
