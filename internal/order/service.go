package order

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrderInput struct {
	Currency string
	Items    []OrderItem
}

type OrderService struct {
	rep OrderRepository
}

func NewOrderService(repository OrderRepository) *OrderService {
	var ord OrderService
	ord.rep = repository

	return &ord
}

func (o *OrderService) FindOrderByID(id string) (Order, error) {
	ord, err := o.rep.FindByID(id)
	if err != nil {
		return Order{}, err
	}
	return ord, nil
}

func (o *OrderService) CreateOrder(customerID string, input CreateOrderInput) (*Order, error) {
	orderID := uuid.NewString()
	now := time.Now()

	order, err := NewOrder(orderID, customerID, input.Currency, input.Items, now)
	if err != nil {
		return nil, err
	}
	err = o.rep.Save(*order)
	if err != nil {
		return nil, err
	}
	return order, nil
}
