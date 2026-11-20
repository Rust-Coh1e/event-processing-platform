package service

import (
	"context"
	"errors"
	"event-processing-platform/internal/domain"
	"fmt"

	"github.com/google/uuid"
)

type EventRepo interface {
	ExistsByID(ctx context.Context, event_id uuid.UUID) (bool, error)
	Save(ctx context.Context, req domain.Event) (int, error)
}

type EventService struct {
	repo EventRepo
}

func NewEventService(repo EventRepo) EventService {
	return EventService{
		repo: repo,
	}
}

var (
	ErrValidation     = errors.New("validation failed")
	ErrEventDuplicate = errors.New("event already exists")
)

func (es EventService) CreateEvent(ctx context.Context, req domain.EventRequest) (int, error) {
	// 1. Базовая валидация полей структуры
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	// 2. Бизнес-валидация (требует обращения к БД)
	exists, err := es.repo.ExistsByID(ctx, req.EventID)
	if err != nil {
		return 0, fmt.Errorf("failed to check event existence: %w", err)
	}
	if exists {
		return 0, ErrEventDuplicate
	}

	// 3. Создание и сохранение
	event := domain.NewEventFromRequest(req)
	var id int
	id, err = es.repo.Save(ctx, event)
	if err != nil {
		return 0, fmt.Errorf("failed to save event: %w", err)
	}

	return id, nil
}
