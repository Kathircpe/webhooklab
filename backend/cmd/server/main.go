package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"webhooklab/internal/config"
	"webhooklab/internal/db"
	"webhooklab/internal/http/handlers"
	httpx "webhooklab/internal/http"
	"webhooklab/internal/repository"
	"webhooklab/internal/service"
	"webhooklab/internal/sse"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}

	repo := repository.New(pool)
	broker := sse.New()
	svc := service.New(cfg, repo, broker)

	health := &handlers.HealthHandler{Pool: pool}
	router := httpx.NewRouter(cfg, svc, health)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: router, ReadHeaderTimeout: 10 * time.Second}

	// cleanup worker
	go svc.CleanupLoop(ctx)

	go func() {
		slog.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancel()
	shutCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	srv.Shutdown(shutCtx)
}
