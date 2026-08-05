package order

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	ord        Order
	saveErr    error
	saveCalled bool
}

func (r *fakeRepository) Save(ctx context.Context, order Order) error {
	r.ord = order
	r.saveCalled = true
	return r.saveErr
}

func (r *fakeRepository) FindByID(ctx context.Context, id string) (Order, error) {
	return Order{}, nil
}

func TestOrderServiceCreateOrderSuccess(t *testing.T) {
	rep := new(fakeRepository)
	service := NewOrderService(rep)

	item := OrderItem{ProductID: "order-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	input := CreateOrderInput{Currency: "RUB", Items: []OrderItem{item}}

	order, err := service.CreateOrder(context.Background(), "1", input)
	if err != nil {
		t.Fatalf("ошибка создания заказа: %v", err)
	}
	if order == nil {
		t.Fatal("пустой заказ")
	}
	if order.CustomerID != "1" {
		t.Errorf("ожидалось айди 1 %v", order.CustomerID)
	}
	if order.Currency != "RUB" {
		t.Errorf("ожидалась rus валюта %v", order.Currency)
	}
	if order.ID != rep.ord.ID {
		t.Errorf("айди заказов не совпадают: %v", order.ID)
	}
	if order.ID == "" {
		t.Error("айди не может быть пустым")
	}
}

func TestOrderServiceCreateOrderSaveError(t *testing.T) {
	expErr := errors.New("ошибка сохранения")

	rep := &fakeRepository{saveErr: expErr}
	service := NewOrderService(rep)

	item := OrderItem{ProductID: "order-2", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	input := CreateOrderInput{Currency: "RUB", Items: []OrderItem{item}}

	order, err := service.CreateOrder(context.Background(), "2", input)
	if order != nil {
		t.Errorf("ожидали пустой заказ: %v", order)
	}
	if !errors.Is(err, expErr) {
		t.Errorf("ожидаемая и полученная ошибки не совпали: %v", err)
	}
}

func TestOrderServiceCreateOrderValidationError(t *testing.T) {
	rep := &fakeRepository{}
	service := NewOrderService(rep)

	item := OrderItem{ProductID: "order-2", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	input := CreateOrderInput{Currency: "RUB", Items: []OrderItem{item}}

	order, err := service.CreateOrder(context.Background(), "", input)
	if order != nil {
		t.Errorf("ожидали пустой заказ: %v", order)
	}
	if !errors.Is(err, ErrCustomerIDRequired) {
		t.Errorf("ожидаемая и полученная ошибки не совпали: %v", err)
	}
	if rep.saveCalled {
		t.Error("репозиторий не должен вызываться при некор данных")
	}
}
