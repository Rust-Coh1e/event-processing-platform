package http

import (
	"context"
	"encoding/json"
	"errors"
	"event-processing-platform/internal/domain"
	"event-processing-platform/internal/service"
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

// var (
// 	ErrValidation     = errors.New("validation failed")
// 	ErrEventDuplicate = errors.New("event already exists")
// )

func (event *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Здесь нужно считать json и вызвать сервис
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req domain.EventRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&req)
	if err != nil {
		http.Error(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := event.service.CreateEvent(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			http.Error(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrValidation) {
			http.Error(w, "Already exist: "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "DB problems", 503)
		return
	}

	data, err := json.Marshal(map[string]int{
		"id": id,
	})
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
