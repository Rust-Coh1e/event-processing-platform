package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"event-processing-platform/internal/config"
	thttp "event-processing-platform/internal/transport/http"
)

func main() {

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		fmt.Print(err)
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", thttp.HealthHandler)

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
