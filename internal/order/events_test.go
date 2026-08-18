package order

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSuccessEvent(t *testing.T) {
	item1 := OrderItem{ProductID: "pr-1", Name: "bruh", UnitPrice: 100, Quantity: 1}
	item2 := OrderItem{ProductID: "pr-2", Name: "bru", UnitPrice: 200, Quantity: 2}
	order := Order{ID: "id-1", CustomerID: "cus-1", Items: []OrderItem{item1, item2},
		Status: StatusCreated, TotalAmount: item1.UnitPrice*int64(item1.Quantity) + item2.UnitPrice*int64(item2.Quantity),
		Currency: "RUB", CreatedAt: time.Date(2026, 12, 1, 1, 1, 1, 0, time.UTC),
		UpdatedAt: time.Date(2026, 12, 1, 1, 2, 1, 0, time.UTC)}

	event, err := NewOrderCreatedEvent(order)

	if err != nil {
		t.Fatalf("ошибка создания ивента %v", err)
	}

	if event.ID == "" {
		t.Fatal("айди ивента пустой")
	}

	if event.AggregateType != "order" {
		t.Fatalf("тип aggregate не order, а %v", event.AggregateType)
	}

	if event.AggregateID != order.ID {
		t.Fatalf("айди aggregate не совпадает, должно быть %v", order.ID)
	}

	if event.EventType != "order.created" {
		t.Fatalf("тип инвента не order.created, а %v", event.EventType)
	}

	if !event.CreatedAt.Equal(order.CreatedAt) {
		t.Fatalf("время создания ивента не то, должно быть %v", order.CreatedAt)
	}

	if event.PublishedAt != nil {
		t.Fatal("published должно быть nil")
	}
	payload := OrderCreatedPayload{}

	err = json.Unmarshal(event.Payload, &payload)
	if err != nil {
		t.Fatalf("ошибка декодирования json %v", err)
	}

	if payload.OrderID != order.ID {
		t.Fatalf("order_id payload не совпадает, должно бытьv %v", order.ID)
	}

	if payload.CustomerID != order.CustomerID {
		t.Fatalf("cus_id payload не совпадает, должно бытьv %v", order.CustomerID)
	}

	if payload.TotalAmount != order.TotalAmount {
		t.Fatalf("total_amount не совпал, должно быть %v", order.TotalAmount)
	}

	if payload.Currency != order.Currency {
		t.Fatalf("currency не совпал, должно быть %v", order.Currency)
	}

	if !payload.CreatedAt.Equal(order.CreatedAt) {
		t.Fatalf("время создания ивента не то, должно быть %v", order.CreatedAt)
	}

	if len(payload.Items) != len(order.Items) {
		t.Fatalf("длины items не совпали, должно быть %v", len(order.Items))
	}

	for i := range payload.Items {
		if payload.Items[i].ProductID != order.Items[i].ProductID {
			t.Fatalf("product_id не совпало, должно быть %v", order.Items[i].ProductID)
		}

		if payload.Items[i].Name != order.Items[i].Name {
			t.Fatalf("name не совпало, должно быть %v", order.Items[i].Name)
		}

		if payload.Items[i].UnitPrice != order.Items[i].UnitPrice {
			t.Fatalf("unit_price не совпало, должно быть %v", order.Items[i].UnitPrice)
		}

		if payload.Items[i].Quantity != order.Items[i].Quantity {
			t.Fatalf("quantity не совпало, должно быть %v", order.Items[i].Quantity)
		}
	}

}
