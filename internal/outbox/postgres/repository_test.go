package postgres

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestRepository(t *testing.T) *Repository {
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
	_, err = pool.Exec(ctx, `TRUNCATE TABLE outbox_events`)
	if err != nil {
		t.Fatalf("очистка не удалась: %v", err)
	}

	return NewRepository(pool)
}
func TestRepositoryFetchUnpublished(t *testing.T) {
	repo := newTestRepository(t)

	query := `INSERT INTO outbox_events(id, aggregate_type, aggregate_id,
	event_type, payload, created_at, published_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.pool.Exec(t.Context(), query, "00000000-0000-0000-0000-000000000001", "type-1", "id-1", "type-1", []byte(`{"order_id": "id-1"}`),
		time.Date(2026, 8, 20, 17, 22, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatalf("ошибка exec: %v", err)
	}

	_, err = repo.pool.Exec(t.Context(), query, "00000000-0000-0000-0000-000000000002", "type-2", "id-2", "type-2", []byte(`{"order_id": "id-2"}`),
		time.Date(2026, 8, 20, 17, 25, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatalf("ошибка exec: %v", err)
	}

	_, err = repo.pool.Exec(t.Context(), query, "00000000-0000-0000-0000-000000000003", "type-3", "id-3", "type-3", []byte(`{"order_id": "id-3"}`),
		time.Date(2026, 8, 20, 17, 20, 0, 0, time.UTC), time.Date(2026, 8, 20, 17, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ошибка exec: %v", err)
	}

	evts, err := repo.FetchUnpublished(t.Context(), 1)
	if err != nil {
		t.Fatalf("error FetchUnpublished %v", err)
	}
	if len(evts) != 1 {
		t.Fatal("длина должна быть 1")
	}
	if evts[0].ID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("id must be event-1, received: %v", evts[0].ID)
	}
	if evts[0].AggregateType != "type-1" {
		t.Fatalf("aggregate_type must be type-1, received: %v", evts[0].AggregateType)
	}
	if evts[0].AggregateID != "id-1" {
		t.Fatalf("aggregate_id must be id-1, received: %v", evts[0].AggregateID)
	}
	if evts[0].EventType != "type-1" {
		t.Fatalf("event_type must be type-1, received: %v", evts[0].EventType)
	}
	if string(evts[0].Payload) != `{"order_id": "id-1"}` {
		t.Fatalf("payload must be payload, received: %v", evts[0].Payload)
	}
	if !time.Time.Equal(evts[0].CreatedAt, time.Date(2026, 8, 20, 17, 22, 0, 0, time.UTC)) {
		t.Fatal("time created_at didnt coincide")
	}
	if evts[0].PublishedAt != nil {
		t.Fatalf("time published_at didnt nil: %v", evts[0].PublishedAt)
	}

	evts2, err := repo.FetchUnpublished(t.Context(), 10)
	if err != nil {
		t.Fatalf("error FetchUnpublished %v", err)
	}
	if len(evts2) != 2 {
		t.Fatal("длина должна быть 2")
	}
	if evts2[0].ID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("first id must be event-1, received: %v", evts2[0].ID)
	}
	if evts2[1].ID != "00000000-0000-0000-0000-000000000002" {
		t.Fatalf("second id must be event-2, received: %v", evts2[1].ID)
	}
}

func TestRepositoryMarkPublished(t *testing.T) {
	repo := newTestRepository(t)

	query := `INSERT INTO outbox_events(id, aggregate_type, aggregate_id,
	event_type, payload, created_at, published_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.pool.Exec(t.Context(), query, "00000000-0000-0000-0000-000000000001", "type-1", "id-1", "type-1", []byte(`{"order_id": "id-1"}`),
		time.Date(2026, 8, 20, 17, 22, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatalf("ошибка exec: %v", err)
	}

	err = repo.MarkPublished(t.Context(), "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("error from MarkPublished %v", err)
	}

	var published bool
	query2 := `SELECT published_at IS NOT NULL FROM outbox_events WHERE id = $1`
	row := repo.pool.QueryRow(t.Context(), query2, "00000000-0000-0000-0000-000000000001")

	err = row.Scan(&published)
	if err != nil {
		t.Fatalf("error from scan %v", err)
	}

	if published == false {
		t.Fatal("published must be true")
	}

	err = repo.MarkPublished(t.Context(), "00000000-0000-0000-0000-000000000001")
	if !errors.Is(err, ErrEventNotPending) {
		t.Fatalf("error must be ErrEventNotPending %v", err)
	}
}
