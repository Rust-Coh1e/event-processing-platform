package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID    uuid.UUID       `json:"event_id"`
	DeviceID   string          `json:"device_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload" validate:"required"`
	CreatedAt  time.Time       `json:"created_at"`
}

type EventRequest struct {
	EventID    uuid.UUID       `json:"event_id" validate:"required"`
	DeviceID   string          `json:"device_id" validate:"required,min=3,max=64"`
	EventType  string          `json:"event_type" validate:"required,oneof=click view purchase"` // пример допустимых типов
	OccurredAt time.Time       `json:"occurred_at" validate:"required"`
	Payload    json.RawMessage `json:"payload" validate:"required"`
}

func NewEventFromRequest(er EventRequest) Event {
	return Event{
		EventID:    er.EventID,
		DeviceID:   er.DeviceID,
		EventType:  er.EventType,
		OccurredAt: er.OccurredAt,
		Payload:    er.Payload,
		CreatedAt:  time.Now(),
	}
}

func (r EventRequest) Validate() error {
	var errs []error

	// 1. Проверка UUID (пустой UUID состоит из одних нулей)
	if r.EventID == uuid.Nil {
		errs = append(errs, errors.New("event_id is required"))
	}

	// 2. Проверка строковых полей (TrimSpace защищает от строк из одних пробелов)
	if strings.TrimSpace(r.DeviceID) == "" {
		errs = append(errs, errors.New("device_id is required"))
	}
	if strings.TrimSpace(r.EventType) == "" {
		errs = append(errs, errors.New("event_type is required"))
	}

	// 3. Проверка времени (IsZero проверяет, что дата вообще была передана)
	if r.OccurredAt.IsZero() {
		errs = append(errs, errors.New("occurred_at is required"))
	} else if r.OccurredAt.After(time.Now().Add(5 * time.Minute)) {
		errs = append(errs, errors.New("occurred_at cannot be in the future"))
	}

	// 4. Проверка JSON Payload
	if len(r.Payload) == 0 || string(r.Payload) == "null" {
		errs = append(errs, errors.New("payload is required"))
	} else if !json.Valid(r.Payload) {
		errs = append(errs, errors.New("payload must be a valid json"))
	}

	// Объединяет все ошибки в одну (вернет nil, если срез errs пуст)
	return errors.Join(errs...)
}

// -- {
// --   "event_id": "fbc6b03d-13ad-4e92-b8a5-a0a8dbde1641",
// --   "device_id": "sensor-001",
// --   "event_type": "temperature_exceeded",
// --   "occurred_at": "2026-11-16T10:00:00Z",
// --   "payload": {
// --     "temperature": 95.5,
// --     "threshold": 90
// --   }
// -- }
