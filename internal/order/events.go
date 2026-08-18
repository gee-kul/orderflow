package order

import (
	"encoding/json"
	"time"

	"github.com/gee-kul/orderflow/internal/outbox"
	"github.com/google/uuid"
)

type OrderCreatedItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	UnitPrice int64  `json:"unit_price"`
	Quantity  int    `json:"quantity"`
}

type OrderCreatedPayload struct {
	OrderID     string             `json:"order_id"`
	CustomerID  string             `json:"customer_id"`
	Items       []OrderCreatedItem `json:"items"`
	TotalAmount int64              `json:"total_amount"`
	Currency    string             `json:"currency"`
	CreatedAt   time.Time          `json:"created_at"`
}

func NewOrderCreatedEvent(order Order) (outbox.Event, error) {
	items := make([]OrderCreatedItem, len(order.Items))

	for i, item := range order.Items {
		items[i].ProductID = item.ProductID
		items[i].Name = item.Name
		items[i].UnitPrice = item.UnitPrice
		items[i].Quantity = item.Quantity
	}

	payload := OrderCreatedPayload{OrderID: order.ID, CustomerID: order.CustomerID, Items: items, TotalAmount: order.TotalAmount, Currency: order.Currency, CreatedAt: order.CreatedAt}

	payloadEncod, err := json.Marshal(payload)
	if err != nil {
		return outbox.Event{}, err
	}

	event := outbox.Event{ID: uuid.NewString(), AggregateType: "order", AggregateID: order.ID,
		EventType: "order.created", Payload: payloadEncod, CreatedAt: order.CreatedAt, PublishedAt: nil}

	return event, nil
}
