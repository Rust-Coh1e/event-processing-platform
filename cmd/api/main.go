package event_processing_platform

import (
	"log"
	"net/http"
	"time"

	thttp "event-processing-platform/internal/transport/http"
)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", thttp.HealthHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server starting on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error server start: %v", err)
	}
}
