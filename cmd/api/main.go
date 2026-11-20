package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-processing-platform/internal/config"
	"event-processing-platform/internal/repository/postgres"
	"event-processing-platform/internal/service"
	thttp "event-processing-platform/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {

	startCtx := context.Background()

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		slog.Error(fmt.Sprintf("Error in cfg: %v", err))
		os.Exit(1)
	}

	pool, err := pgxpool.New(startCtx, cfg.Dsn)
	if err != nil {
		slog.Error("failed to create db pool", "err", err)
		os.Exit(1)
	}
	// Гарантируем закрытие соединений при выходе из main
	defer pool.Close()

	pingCtx, close := context.WithTimeout(startCtx, time.Second*10)
	defer close()

	if err := pool.Ping(pingCtx); err != nil {
		slog.Error("failed to ping db", "err", err)
		os.Exit(1)
	}

	eventRepo := postgres.NewEventRepo(pool)
	eventService := service.NewEventService(eventRepo)
	eventHandler := thttp.NewEventHandler(eventService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", thttp.HealthHandler)
	mux.HandleFunc("POST /api/v1/events", eventHandler.CreateEvent)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err = runServer(ctx, server, cfg.ShutdownTOut)
	if err != nil {
		slog.Error(fmt.Sprintf("Error in server: %v", err), "addr", server.Addr)
	}

	// Закрыть соединения с внешними системами

}
