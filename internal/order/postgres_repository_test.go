package order

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestPostgresRepository(t *testing.T) *PostgresOrderRepository {
	t.Helper()
	ctx := context.Background()
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
	_, err = pool.Exec(ctx, `TRUNCATE TABLE order_items, orders, outbox_events`)
	if err != nil {
		t.Fatalf("очистка не удалась: %v", err)
	}

	return NewPostgresOrderRepository(pool)
}

func TestPostgresOrderRepositorySaveAndRead(t *testing.T) {
	repo := newTestPostgresRepository(t)

	item1 := OrderItem{ProductID: "product-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	item2 := OrderItem{ProductID: "product-2", Name: "bru", UnitPrice: 100000, Quantity: 2}
	order := Order{ID: "order-1", CustomerID: "cust", Items: []OrderItem{item1, item2}, Status: StatusCreated,
		TotalAmount: item1.UnitPrice*int64(item1.Quantity) + item2.UnitPrice*int64(item2.Quantity),
		Currency:    "RUB", CreatedAt: time.Date(2026, 12, 1, 1, 1, 1, 0, time.UTC),
		UpdatedAt: time.Date(2026, 12, 1, 1, 2, 1, 0, time.UTC)}

	err := repo.Save(t.Context(), order)
	if err != nil {
		t.Fatalf("ошибка сохранения заказа:%v", err)
	}
	orderByID, err := repo.FindByID(t.Context(), order.ID)
	if err != nil {
		t.Fatalf("не удалось найти заказ по айди:%v", err)
	}

	if orderByID.ID != order.ID {
		t.Fatalf("айди не совпадает, должно быть:%v, вывелось:%v", order.ID, orderByID.ID)
	}
	if orderByID.CustomerID != order.CustomerID {
		t.Fatalf("customer айди не совпадает, должно быть:%v, вывелось:%v", order.CustomerID, orderByID.CustomerID)
	}
	if orderByID.Status != order.Status {
		t.Fatalf("status не совпадает, должно быть:%v, вывелось:%v", order.Status, orderByID.Status)
	}
	if orderByID.TotalAmount != order.TotalAmount {
		t.Fatalf("total amount не совпадает, должно быть:%v, вывелось:%v", order.TotalAmount, orderByID.TotalAmount)
	}
	if orderByID.Currency != order.Currency {
		t.Fatalf("currency не совпадает, должно быть:%v, вывелось:%v", order.Currency, orderByID.Currency)
	}
	if !time.Time.Equal(orderByID.CreatedAt, order.CreatedAt) {
		t.Fatalf("created at не совпадает, должно быть:%v, вывелось:%v", order.CreatedAt, orderByID.CreatedAt)
	}
	if !time.Time.Equal(orderByID.UpdatedAt, order.UpdatedAt) {
		t.Fatalf("updated at не совпадает, должно быть:%v, вывелось:%v", order.UpdatedAt, orderByID.UpdatedAt)
	}
	if !reflect.DeepEqual(order.Items, orderByID.Items) {
		t.Fatalf("items не совпадают, должно быть:%v, вывелось:%v", order.Items, orderByID.Items)
	}
}

func TestPostgresOrderRepositoryNotExistID(t *testing.T) {
	repo := newTestPostgresRepository(t)
	_, err := repo.FindByID(t.Context(), "666")
	if err == nil {
		t.Fatal("должна была вернутся ошибка")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("неожиданная ошибка:%v", err)
	}
}

func TestPostgresOrderRepositoryRepeatSave(t *testing.T) {
	repo := newTestPostgresRepository(t)

	item1 := OrderItem{ProductID: "product-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	item2 := OrderItem{ProductID: "product-2", Name: "bru", UnitPrice: 100000, Quantity: 2}
	item3 := OrderItem{ProductID: "product-3", Name: "br", UnitPrice: 50000, Quantity: 3}
	order := Order{ID: "order-1", CustomerID: "cust", Items: []OrderItem{item1, item2, item3}, Status: StatusCreated,
		TotalAmount: item1.UnitPrice*int64(item1.Quantity) + item2.UnitPrice*int64(item2.Quantity) + item3.UnitPrice*int64(item3.Quantity),
		Currency:    "RUB", CreatedAt: time.Date(2026, 12, 1, 1, 1, 1, 0, time.UTC),
		UpdatedAt: time.Date(2026, 12, 1, 1, 2, 1, 0, time.UTC)}

	err := repo.Save(t.Context(), order)
	if err != nil {
		t.Fatalf("ошибка сохранения заказа:%v", err)
	}

	orderNew := Order{ID: "order-1", CustomerID: "cust-new", Items: []OrderItem{item2, item1}, Status: StatusProcessing,
		TotalAmount: item1.UnitPrice*int64(item1.Quantity) + item2.UnitPrice*int64(item2.Quantity),
		Currency:    "USD", CreatedAt: time.Date(2026, 12, 1, 2, 1, 1, 0, time.UTC),
		UpdatedAt: time.Date(2026, 12, 1, 2, 2, 1, 0, time.UTC)}

	err = repo.Save(t.Context(), orderNew)
	if err != nil {
		t.Fatalf("ошибка сохранения заказа:%v", err)
	}

	orderByID, err := repo.FindByID(t.Context(), orderNew.ID)
	if err != nil {
		t.Fatalf("не удалось найти заказ по айди:%v", err)
	}

	if orderByID.Status != orderNew.Status {
		t.Fatalf("status не совпадает, должно быть:%v, вывелось:%v", orderNew.Status, orderByID.Status)
	}
	if orderByID.TotalAmount != orderNew.TotalAmount {
		t.Fatalf("total amount не совпадает, должно быть:%v, вывелось:%v", orderNew.TotalAmount, orderByID.TotalAmount)
	}
	if !time.Time.Equal(orderByID.CreatedAt, orderNew.CreatedAt) {
		t.Fatalf("created at не совпадает, должно быть:%v, вывелось:%v", orderNew.CreatedAt, orderByID.CreatedAt)
	}
	if !time.Time.Equal(orderByID.UpdatedAt, orderNew.UpdatedAt) {
		t.Fatalf("updated at не совпадает, должно быть:%v, вывелось:%v", orderNew.UpdatedAt, orderByID.UpdatedAt)
	}
	if orderByID.CustomerID != orderNew.CustomerID {
		t.Fatalf("customer айди не совпадает, должно быть:%v, вывелось:%v", orderNew.CustomerID, orderByID.CustomerID)
	}
	if orderByID.Currency != orderNew.Currency {
		t.Fatalf("currency не совпадает, должно быть:%v, вывелось:%v", orderNew.Currency, orderByID.Currency)
	}
	if !reflect.DeepEqual(orderNew.Items, orderByID.Items) {
		t.Fatalf("items не совпадают, должно быть:%v, вывелось:%v", orderNew.Items, orderByID.Items)
	}
}

func TestPostgresOrderRepositorySaveWithEvent(t *testing.T) {
	repo := newTestPostgresRepository(t)

	item1 := OrderItem{ProductID: "product-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	item2 := OrderItem{ProductID: "product-2", Name: "bru", UnitPrice: 100000, Quantity: 2}
	order := Order{ID: "order-1", CustomerID: "cust", Items: []OrderItem{item1, item2}, Status: StatusCreated,
		TotalAmount: item1.UnitPrice*int64(item1.Quantity) + item2.UnitPrice*int64(item2.Quantity),
		Currency:    "RUB", CreatedAt: time.Date(2026, 12, 1, 1, 1, 1, 0, time.UTC),
		UpdatedAt: time.Date(2026, 12, 1, 1, 2, 1, 0, time.UTC)}

	evt, err := NewOrderCreatedEvent(order)
	if err != nil {
		t.Fatalf("ошибка создания нового ивента %v", err)
	}

	err = repo.SaveWithEvent(t.Context(), order, evt)
	if err != nil {
		t.Fatalf("ошибка сохранения заказа с ивентом:%v", err)
	}

	_, err = repo.FindByID(t.Context(), order.ID)
	if err != nil {
		t.Fatalf("не удалось найти заказ по айди:%v", err)
	}

	query := `SELECT aggregate_type, aggregate_id, event_type, published_at
	FROM outbox_events WHERE id = $1`

	row := repo.pool.QueryRow(t.Context(), query, evt.ID)
	var aggregateType string
	var aggregateID string
	var eventType string
	var publishedAt *time.Time
	err = row.Scan(&aggregateType, &aggregateID, &eventType, &publishedAt)
	if err != nil {
		t.Fatalf("ошибка скана %v", err)
	}

	if evt.AggregateType != aggregateType {
		t.Errorf("aggregate_type  не совпало, должно быть %v", evt.AggregateType)
	}

	if evt.AggregateID != aggregateID {
		t.Errorf("aggregate_id  не совпало, должно быть %v", evt.AggregateID)
	}

	if evt.EventType != eventType {
		t.Errorf("event_type  не совпало, должно быть %v", evt.EventType)
	}

	if publishedAt != nil {
		t.Error("published_at должно быть nil")
	}
}

func TestPostgresOrderRepositorySaveWithEventRollback(t *testing.T) {
	repo := newTestPostgresRepository(t)

	item1 := OrderItem{ProductID: "product-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	order := Order{ID: "order-1", CustomerID: "cust", Items: []OrderItem{item1}, Status: StatusCreated,
		TotalAmount: item1.UnitPrice * int64(item1.Quantity),
		Currency:    "RUB", CreatedAt: time.Date(2026, 12, 1, 1, 1, 1, 0, time.UTC),
		UpdatedAt: time.Date(2026, 12, 1, 1, 2, 1, 0, time.UTC)}

	evt, err := NewOrderCreatedEvent(order)
	if err != nil {
		t.Fatalf("ошибка создания нового ивента %v", err)
	}

	evt.ID = "incorrect id"

	err = repo.SaveWithEvent(t.Context(), order, evt)
	if err == nil {
		t.Fatalf("должна быть ошибка сохранения заказа с неверным айди:%v", err)
	}

	_, err = repo.FindByID(t.Context(), order.ID)
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("должна быть ошибка заказ не найден, а пришла ошибка :%v", err)
	}
}
