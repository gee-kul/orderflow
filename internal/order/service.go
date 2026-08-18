package order

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateOrderInput struct {
	Currency string
	Items    []OrderItem
}

type OrderService struct {
	rep   OrderRepository
	eventSaver OrderEventSaver
}

func NewOrderService(repository OrderRepository, event OrderEventSaver) *OrderService {
	var ord OrderService
	ord.rep = repository
	ord.eventSaver = event

	return &ord
}

func (o *OrderService) FindOrderByID(ctx context.Context, id string) (Order, error) {
	ord, err := o.rep.FindByID(ctx, id)
	if err != nil {
		return Order{}, err
	}
	return ord, nil
}

func (o *OrderService) CreateOrder(ctx context.Context, customerID string, input CreateOrderInput) (*Order, error) {
	orderID := uuid.NewString()
	now := time.Now()

	order, err := NewOrder(orderID, customerID, input.Currency, input.Items, now)
	if err != nil {
		return nil, err
	}

	event, err := NewOrderCreatedEvent(*order)
	if err != nil {
		return nil, err
	}

	err = o.eventSaver.SaveWithEvent(ctx, *order, event)
	if err != nil {
		return nil, err
	}

	return order, nil
}
