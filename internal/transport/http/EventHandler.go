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
	defer r.Body.Close()

	req, err := domain.DecodeAndValidate[domain.EventRequest](w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := event.service.CreateEvent(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrEventDuplicate):
			http.Error(w, "Already exist: "+err.Error(), http.StatusConflict)
		default:
			http.Error(w, "DB problems", http.StatusServiceUnavailable)
		}
		return
	}

	data, err := json.Marshal(map[string]int{"id": id})
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(data)
}
