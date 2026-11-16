package main

import (
	"log"
	"net/http"
	"time"

	"event-processing-platform/internal/config"
	thttp "event-processing-platform/internal/transport/http"
)

func main() {

	cfg := config.LoadConfig()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", thttp.HealthHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on http://localhost:%v", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error server start: %v", err)
	}
}
