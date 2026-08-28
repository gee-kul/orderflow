package postgres

import (
	"context"
	"fmt"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	var handler Handler
	handler.pool = pool
	return &handler
}

func (h *Handler) Handle(ctx context.Context, evt event.Event) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error from begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO processed_events (event_id) VALUES ($1)
	ON CONFLICT (event_id) DO NOTHING`

	tag, err := tx.Exec(ctx, query, evt.ID)
	if err != nil {
		return fmt.Errorf("error from exec: %w", err)
	}

	rows := tag.RowsAffected()
	if rows == 0 {
		err := tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("error from commit duplicate event: %w", err)
		}
		return nil
	}

	query2 := `INSERT INTO order_event_stats (event_type, event_count)
	VALUES ($1, 1) ON CONFLICT (event_type) DO UPDATE
	SET event_count = order_event_stats.event_count + 1, updated_at = NOW()`

	_, err = tx.Exec(ctx, query2, evt.EventType)
	if err != nil {
		return fmt.Errorf("error from exec: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error from commit event statistics: %w", err)
	}
	return nil
}
