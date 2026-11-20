package postgres

import (
	"context"
	"event-processing-platform/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepo(pl *pgxpool.Pool) EventRepo {
	return EventRepo{
		pool: pl,
	}
}

func (ER EventRepo) ExistsByID(ctx context.Context, event_id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE event_id = $1)`

	var exists bool

	err := ER.pool.QueryRow(ctx, query, event_id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil

}

func (ER EventRepo) Save(ctx context.Context, req domain.Event) (int, bool, error) {
	query := `INSERT INTO events (event_id, device_id, event_type, occurred_at, payload)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (event_id) DO UPDATE
				SET event_id = EXCLUDED.event_id
				RETURNING id, (xmax = 0) AS is_created`
	var id int
	var created bool

	err := ER.pool.QueryRow(
		ctx,
		query,
		req.EventID,
		req.DeviceID,
		req.EventType,
		req.OccurredAt,
		req.Payload,
		req.CreatedAt,
	).Scan(&id, &created)

	if err != nil {
		return 0, false, err
	}
	return id, created, nil
}
