package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func runServer(
	ctx context.Context,
	server *http.Server,
	shutdownTimeout time.Duration,
) error {

	errChan := make(chan error, 1)

	log.Printf("Server starting on http://%v", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			// stop()
			errChan <- fmt.Errorf("Error server start:%w", err)
		}
	}()
	select {
	case <-ctx.Done():
		// Получение сигнала отмены
		log.Printf("Get os cancel signal. Server stoping...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown timed out (%v), forcing server close...", err)

			// Явно принудительно разрываем все активные соединения и освобождаем сокеты
			if closeErr := server.Close(); closeErr != nil {
				return fmt.Errorf("Error when closing server: %w — %w", closeErr, err)
			} else {
				return err
			}

		} else {
			return nil
		}
	case err := <-errChan:
		//Сервер не стратанул можно без GS
		return err
	}

}
