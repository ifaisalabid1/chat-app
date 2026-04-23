package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ifaisalabid1/chat-app/internal/config"
	"github.com/ifaisalabid1/chat-app/internal/delivery/ws"
	"github.com/ifaisalabid1/chat-app/internal/repository"
	"github.com/ifaisalabid1/chat-app/internal/storage"
	"github.com/ifaisalabid1/chat-app/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := storage.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to initialize postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to postgres")

	redisClient, err := storage.NewRedisClient(ctx, cfg.RedisURL, "")
	if err != nil {
		logger.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	logger.Info("connected to redis")

	chatRepo := repository.NewPostgresChatRepo(pool)
	chatUsecase := usecase.NewChatUsecase(chatRepo, 5*time.Second)

	hub := ws.NewHub(redisClient, chatUsecase)

	go hub.Run(context.Background())

	wsHandler := ws.NewHandler(hub)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	})

	r.Get("/ws", wsHandler.ServeWS)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server...", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	logger.Info("starting server", "port", "8080")

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	if err := <-shutdownError; err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped successfully")
}
