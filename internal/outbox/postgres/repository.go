package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	repo := Repository{}
	repo.pool = pool
	return &repo
}

var ErrEventNotPending = errors.New("событие не ожидает публикации")

func (repo *Repository) FetchUnpublished(ctx context.Context, limit int) ([]event.Event, error) {
	evts := []event.Event{}
	query := `SELECT id, aggregate_type, aggregate_id, event_type, payload,
	created_at, published_at FROM outbox_events
	WHERE published_at IS NULL ORDER BY created_at ASC, id ASC LIMIT $1`

	rows, err := repo.pool.Query(ctx, query, limit)
	if err != nil {
		return []event.Event{}, fmt.Errorf("ошибка при получении столбцов: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		evt := event.Event{}
		err = rows.Scan(&evt.ID, &evt.AggregateType, &evt.AggregateID,
			&evt.EventType, &evt.Payload, &evt.CreatedAt, &evt.PublishedAt)
		if err != nil {
			return []event.Event{}, fmt.Errorf("ошибка при сканировании столбцов: %w", err)
		}

		evts = append(evts, evt)
	}

	err = rows.Err()
	if err != nil {
		return []event.Event{}, fmt.Errorf("ошибка rows: %w", err)
	}

	return evts, nil
}

func (repo *Repository) MarkPublished(ctx context.Context, id string) error {
	query := `UPDATE outbox_events SET published_at = NOW()
	WHERE id = $1 AND published_at IS NULL`

	tag, err := repo.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка exec: %w", err)
	}

	count := tag.RowsAffected()
	if count == 0 {
		return fmt.Errorf("count не может быть нулевым %w", ErrEventNotPending)
	}

	return nil
}
