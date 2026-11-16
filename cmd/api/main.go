package main

import (
	"fmt"
	"log"
	"net/http"
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

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", thttp.HealthHandler)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server starting on http://%v", cfg.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error server start: %v", err)
	}
}
