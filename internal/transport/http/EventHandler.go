package http

import (
	"context"
	"encoding/json"
	"errors"
	"event-processing-platform/internal/domain"
	"net/http"
)

type EventService interface {
	CreateEvent(ctx context.Context, req domain.EventRequest) (int, error)
}

type EventHandler struct {
	service EventService
}

func NewEventHandler(service EventService) EventHandler {
	return EventHandler{
		service: service,
	}
}

var (
	ErrValidation     = errors.New("validation failed")
	ErrEventDuplicate = errors.New("event already exists")
)

func (event *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Здесь нужно считать json и вызвать сервис
	defer r.Body.Close()

	var req domain.EventRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := event.service.CreateEvent(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			http.Error(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrEventDuplicate) {
			http.Error(w, "Already exist: "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "DB problems", 503)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err1 := json.NewEncoder(w).Encode(map[string]int{
		"id": id,
	})
	if err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
