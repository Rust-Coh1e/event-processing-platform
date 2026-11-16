package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"event-processing-platform/internal/config"
	thttp "event-processing-platform/internal/transport/http"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Print(err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", thttp.HealthHandler)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on http://%v", cfg.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Error server start")
			stop()
			return
		}
	}()

	<-ctx.Done()
	// Получение сигнала отмены

	log.Printf("Get os cancel signal. Server stoping...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTOut)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown timed out (%v), forcing server close...", err)

		// Явно принудительно разрываем все активные соединения и освобождаем сокеты
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("Error when closing server")
		}

	} else {
		log.Printf("Server gracefully shutted down")
	}

	// Закрыть соединения с внешними системами

}
