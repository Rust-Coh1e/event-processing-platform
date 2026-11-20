package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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

type Validator interface {
	Validate() error
}

const (
	maxDeviceIDLen  = 128
	maxEventTypeLen = 64
)

func (r EventRequest) Validate() error {
	var errs []error

	// 1. Проверка UUID
	if r.EventID == uuid.Nil {
		errs = append(errs, errors.New("event_id is required"))
	}

	// 2. Проверка строковых полей (обязательность + максимальная длина в символах)
	deviceID := strings.TrimSpace(r.DeviceID)
	if deviceID == "" {
		errs = append(errs, errors.New("device_id is required"))
	} else if utf8.RuneCountInString(deviceID) > maxDeviceIDLen {
		errs = append(errs, fmt.Errorf("device_id must be <= %d characters", maxDeviceIDLen))
	}

	eventType := strings.TrimSpace(r.EventType)
	if eventType == "" {
		errs = append(errs, errors.New("event_type is required"))
	} else if utf8.RuneCountInString(eventType) > maxEventTypeLen {
		errs = append(errs, fmt.Errorf("event_type must be <= %d characters", maxEventTypeLen))
	}

	// 3. Проверка времени
	if r.OccurredAt.IsZero() {
		errs = append(errs, errors.New("occurred_at is required"))
	} else if r.OccurredAt.After(time.Now().Add(5 * time.Minute)) {
		errs = append(errs, errors.New("occurred_at cannot be in the future"))
	}

	// 4. Проверка JSON Payload (только JSON-объект "{...}", не массив "[...]" и не null)
	trimmedPayload := bytes.TrimSpace(r.Payload)
	if len(trimmedPayload) == 0 || bytes.Equal(trimmedPayload, []byte("null")) {
		errs = append(errs, errors.New("payload is required"))
	} else if !json.Valid(trimmedPayload) {
		errs = append(errs, errors.New("payload must be a valid json"))
	} else if trimmedPayload[0] != '{' {
		errs = append(errs, errors.New("payload must be a json object"))
	}

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

func DecodeAndValidate[T Validator](w http.ResponseWriter, r *http.Request) (T, error) {
	var req T
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		return req, fmt.Errorf("неверный формат JSON: %w", err)
	}

	// Защита от мусора после первого JSON-объекта (например, "{}{}" или "{} 42")
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return req, errors.New("тело запроса должно содержать только один JSON-объект")
	}

	if err := req.Validate(); err != nil {
		return req, err
	}

	return req, nil
}

func puk() {
	return
}
