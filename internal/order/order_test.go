package order

import (
	"testing"
	"time"
)

func TestNewOrderSuccess(t *testing.T) {

	id := " 1742 "
	customerID := " gee "
	currency := " rub "
	now := time.Date(2026, 8, 16, 18, 1, 0, 0, time.UTC)

	items := make([]OrderItem, 2)

	items[0].ProductID = " 8539 "
	items[0].Name = " 1 "
	items[0].Quantity = 1
	items[0].UnitPrice = 50000

	items[1].ProductID = " 8540 "
	items[1].Name = " 2 "
	items[1].Quantity = 2
	items[1].UnitPrice = 25000

	result, err := NewOrder(id, customerID, currency, items, now)

	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("ожидали result")
	}
	if result.ID != "1742" {
		t.Errorf("ожидалось 1742, %v", result.ID)
	}
	if result.CustomerID != "gee" {
		t.Errorf("ожидалось gee, %v", result.CustomerID)
	}
	if result.Currency != "RUB" {
		t.Errorf("ожидалось RUB, %v", result.Currency)
	}
	if len(result.Items) != 2 {
		t.Fatal("ожидали длину айтемов 2")
	}
	if result.Items[0].ProductID != "8539" {
		t.Errorf("ожидалось 8539 у первой позиции, %v", result.Items[0].ProductID)
	}
	if result.Items[0].Name != "1" {
		t.Errorf("ожидалось 1 у первой позиции, %v", result.Items[0].Name)
	}
	if result.Items[1].ProductID != "8540" {
		t.Errorf("ожидалось 8540 у второй позиции, %v", result.Items[1].ProductID)
	}
	if result.Items[1].Name != "2" {
		t.Errorf("ожидалось 2 у второй позиции, %v", result.Items[1].Name)
	}
	if result.TotalAmount != 100000 {
		t.Errorf("ожидалась сумма 100000, %v", result.TotalAmount)
	}
	if result.Status != StatusCreated {
		t.Errorf("ожидался StatusCreated, %v", result.Status)
	}
	if result.Items[0].Quantity != 1 {
		t.Errorf("ожидался 1 единица, %v", result.Items[0].Quantity)
	}
	if result.Items[0].UnitPrice != 50000 {
		t.Errorf("ожидалась cумма 50000, %v", result.Items[0].UnitPrice)
	}
	if result.Items[1].Quantity != 2 {
		t.Errorf("ожидалась 2 единица, %v", result.Items[1].Quantity)
	}
	if result.Items[1].UnitPrice != 25000 {
		t.Errorf("ожидалась cумма 25000, %v", result.Items[1].UnitPrice)
	}
	if !now.Equal(result.CreatedAt) {
		t.Errorf("ожидалось одинаковое время у CreatedAt, %v", result.CreatedAt)
	}
	if !now.Equal(result.UpdatedAt) {
		t.Errorf("ожидалось одинаковое время у UpdatedAt, %v", result.UpdatedAt)
	}
	items[0].ProductID = " 8567 "
	if result.Items[0].ProductID != "8539" {
		t.Errorf("ожидалось 8539 у первой позиции, %v", result.Items[0].ProductID)
	}

}
