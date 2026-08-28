package postgres

import (
	"os"
	"testing"

	"github.com/gee-kul/orderflow/internal/event"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	ctx := t.Context()
	str := os.Getenv("ORDERFLOW_TEST_DATABASE_URL")
	if str == "" {
		t.Skip("ORDERFLOW_TEST_DATABASE_URL не задана")
	}
	pool, err := pgxpool.New(ctx, str)
	if err != nil {
		t.Fatalf("ошибка пула соединений: %v", err)
	}
	t.Cleanup(pool.Close)

	err = pool.Ping(ctx)
	if err != nil {
		t.Fatalf("ошибка пинга: %v", err)
	}
	_, err = pool.Exec(ctx, `TRUNCATE TABLE processed_events, order_event_stats`)
	if err != nil {
		t.Fatalf("очистка не удалась: %v", err)
	}

	return NewHandler(pool)
}

func TestHandlerHandleNewEvent(t *testing.T) {
	evt := event.Event{ID: uuid.NewString(), AggregateType: "order",
		AggregateID: "order-1", EventType: "order.created"}

	handler := newTestHandler(t)
	err := handler.Handle(t.Context(), evt)
	if err != nil {
		t.Fatalf("error from handle: %v", err)
	}

	query := `SELECT COUNT(*) FROM processed_events WHERE event_id = $1`

	row := handler.pool.QueryRow(t.Context(), query, evt.ID)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("error from scan: %v", err)
	}
	if count != 1 {
		t.Errorf("count must be 1, but received %v", count)
	}

	query2 := `SELECT event_count FROM order_event_stats WHERE event_type = $1`

	row2 := handler.pool.QueryRow(t.Context(), query2, evt.EventType)
	var count2 int64
	err = row2.Scan(&count2)
	if err != nil {
		t.Fatalf("error from scan: %v", err)
	}
	if count2 != 1 {
		t.Errorf("count must be 1, but received %v", count2)
	}
}

func TestHandlerHandleDuplicateEvent(t *testing.T) {
	evt := event.Event{ID: uuid.NewString(), AggregateType: "order",
		AggregateID: "order-1", EventType: "order.created"}

	handler := newTestHandler(t)
	err := handler.Handle(t.Context(), evt)
	if err != nil {
		t.Fatalf("error from handle: %v", err)
	}
	err = handler.Handle(t.Context(), evt)
	if err != nil {
		t.Fatalf("error from handle: %v", err)
	}

	query := `SELECT COUNT(*) FROM processed_events WHERE event_id = $1`

	row := handler.pool.QueryRow(t.Context(), query, evt.ID)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("error from scan: %v", err)
	}
	if count != 1 {
		t.Fatalf("count must be 1, but received %v", count)
	}

	query2 := `SELECT event_count FROM order_event_stats WHERE event_type = $1`

	row2 := handler.pool.QueryRow(t.Context(), query2, evt.EventType)
	var count2 int64
	err = row2.Scan(&count2)
	if err != nil {
		t.Fatalf("error from scan: %v", err)
	}
	if count2 != 1 {
		t.Fatalf("count must be 1, but received %v", count2)
	}
}

func TestHandlerHandleDifferentEvents(t *testing.T) {
	evt := event.Event{ID: uuid.NewString(), AggregateType: "order",
		AggregateID: "order-1", EventType: "order.created"}

	evt2 := event.Event{ID: uuid.NewString(), AggregateType: "order",
		AggregateID: "order-2", EventType: "order.created"}

	handler := newTestHandler(t)
	err := handler.Handle(t.Context(), evt)
	if err != nil {
		t.Fatalf("error from handle evt: %v", err)
	}
	err = handler.Handle(t.Context(), evt2)
	if err != nil {
		t.Fatalf("error from handle evt2: %v", err)
	}

	query := `SELECT COUNT(*) FROM processed_events`

	row := handler.pool.QueryRow(t.Context(), query)
	var count int64
	err = row.Scan(&count)
	if err != nil {
		t.Fatalf("error from scan: %v", err)
	}
	if count != 2 {
		t.Errorf("count must be 2, but received %v", count)
	}

	query2 := `SELECT event_count FROM order_event_stats WHERE event_type = $1`

	row2 := handler.pool.QueryRow(t.Context(), query2, evt.EventType)
	var count2 int64
	err = row2.Scan(&count2)
	if err != nil {
		t.Fatalf("error from scan: %v", err)
	}
	if count2 != 2 {
		t.Errorf("count must be 2, but received %v", count2)
	}
}
