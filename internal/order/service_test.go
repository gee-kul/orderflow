package order

import (
	"context"
	"errors"
	"testing"

	"github.com/gee-kul/orderflow/internal/event"
)

type fakeRepository struct {
	ord                 Order
	saveErr             error
	saveCalled          bool
	event               event.Event
	saveWithEventCalled bool
}

func (r *fakeRepository) Save(ctx context.Context, order Order) error {
	r.ord = order
	r.saveCalled = true
	return r.saveErr
}

func (r *fakeRepository) SaveWithEvent(ctx context.Context, order Order, event event.Event) error {
	r.ord = order
	r.event = event
	r.saveWithEventCalled = true
	return r.saveErr
}

func (r *fakeRepository) FindByID(ctx context.Context, id string) (Order, error) {
	return Order{}, nil
}

func TestOrderServiceCreateOrderSuccess(t *testing.T) {
	rep := new(fakeRepository)
	service := NewOrderService(rep, rep)

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

	if rep.saveWithEventCalled == false {
		t.Fatal("saveWithEventCalled должен быть true")
	}

	if rep.saveCalled == true {
		t.Fatal("saveCalled должен быть false")
	}

	if rep.event.AggregateID != order.ID {
		t.Errorf("айди ивента не совпало, должно быть %v", order.ID)
	}

	if rep.event.EventType != "order.created" {
		t.Errorf("event_type не order.created а %v", rep.event.EventType)
	}
}

func TestOrderServiceCreateOrderSaveError(t *testing.T) {
	expErr := errors.New("ошибка сохранения")

	rep := &fakeRepository{saveErr: expErr}
	service := NewOrderService(rep, rep)

	item := OrderItem{ProductID: "order-2", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	input := CreateOrderInput{Currency: "RUB", Items: []OrderItem{item}}

	order, err := service.CreateOrder(context.Background(), "2", input)
	if order != nil {
		t.Errorf("ожидали пустой заказ: %v", order)
	}
	if !errors.Is(err, expErr) {
		t.Errorf("ожидаемая и полученная ошибки не совпали: %v", err)
	}

	if rep.saveWithEventCalled == false {
		t.Fatal("ошибка возвращается не из saveWithEventCalled")
	}
}

func TestOrderServiceCreateOrderValidationError(t *testing.T) {
	rep := &fakeRepository{}
	service := NewOrderService(rep, rep)

	item := OrderItem{ProductID: "order-2", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	input := CreateOrderInput{Currency: "RUB", Items: []OrderItem{item}}

	order, err := service.CreateOrder(context.Background(), "", input)
	if order != nil {
		t.Errorf("ожидали пустой заказ: %v", order)
	}
	if !errors.Is(err, ErrCustomerIDRequired) {
		t.Errorf("ожидаемая и полученная ошибки не совпали: %v", err)
	}

	if rep.saveCalled != false || rep.saveWithEventCalled != false {
		t.Fatal("оба флага save должны быть false")
	}
}
