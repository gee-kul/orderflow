package order

import (
	"errors"
	"math"
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

type testCase struct {
	name       string
	id         string
	customerID string
	currency   string
	orderItems []OrderItem
	err        error
}

func TestNewOrderValidation(t *testing.T) {
	cases := []testCase{
		testCase{
			name:       "id только из пробелов",
			id:         " ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: 82742,
					Quantity:  423,
				},
			},
			err: ErrOrderIDRequired,
		},
		testCase{
			name:       "customer id только из пробелов",
			id:         " 87",
			customerID: " ",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: 82742,
					Quantity:  423,
				},
			},
			err: ErrCustomerIDRequired,
		},
		testCase{
			name:       "currency только из пробелов",
			id:         "78 ",
			customerID: "64",
			currency:   " ",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: 82742,
					Quantity:  423,
				},
			},
			err: ErrCurrencyRequired,
		},
		testCase{
			name:       "слайс позиций пустой",
			id:         " 67",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{},
			err:        ErrItemsRequired,
		},

		testCase{
			name:       "product id пустой",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "",
					Name:      "hvgfk",
					UnitPrice: 82742,
					Quantity:  423,
				},
			},
			err: ErrProductIDRequired,
		},

		testCase{
			name:       "name пустой",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "",
					UnitPrice: 82742,
					Quantity:  423,
				},
			},
			err: ErrProductNameRequired,
		},
		testCase{
			name:       "unitPrice нулевой",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: 0,
					Quantity:  423,
				},
			},
			err: ErrUnitPriceInvalid,
		},

		testCase{
			name:       "unitPrice меньше нуля",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: -1,
					Quantity:  423,
				},
			},
			err: ErrUnitPriceInvalid,
		},

		testCase{
			name:       "quantity нулевой",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: 82742,
					Quantity:  0,
				},
			},
			err: ErrQuantityInvalid,
		},

		testCase{
			name:       "quantity меньше нуля",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: 82742,
					Quantity:  -1,
				},
			},
			err: ErrQuantityInvalid,
		},

		testCase{
			name:       "умножение цены на колво переполнено",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: math.MaxInt64,
					Quantity:  2,
				},
			},
			err: ErrTotalAmountOverflow,
		},

		testCase{
			name:       "общая сумма при сложении переполнена",
			id:         "78 ",
			customerID: "64",
			currency:   "RUB",
			orderItems: []OrderItem{
				{
					ProductID: "r8r",
					Name:      "hvgfk",
					UnitPrice: math.MaxInt64,
					Quantity:  1,
				},
				{
					ProductID: "r3vu",
					Name:      "hvgffbk",
					UnitPrice: 82742,
					Quantity:  1,
				},
			},
			err: ErrTotalAmountOverflow,
		},
	}
	now := time.Date(2026, 8, 16, 18, 1, 0, 0, time.UTC)
	for _, caseN := range cases {
		t.Run(caseN.name, func(t *testing.T) {
			order, err := NewOrder(caseN.id, caseN.customerID, caseN.currency, caseN.orderItems, now)
			if err == nil {
				t.Fatal("ожидалась ошибка, вернулся нил")
			}
			if !errors.Is(err, caseN.err) {
				t.Errorf("ожидали: %v, получили:%v ", caseN.err, err)
			}
			if order != nil {
				t.Error("некорректный заказ был создан")
			}

		})

	}
}

func TestOrderChangeStatusAllowed(t *testing.T){
	createdAt := time.Date(2026, 8, 16, 18, 1, 0, 0, time.UTC)
	changedAt := time.Date(2026, 8, 16, 18, 2, 0, 0, time.UTC)
	order := &Order{Status: StatusCreated, CreatedAt: createdAt, UpdatedAt: createdAt}

	err := order.ChangeStatus(StatusConfirmed, changedAt)
	if err != nil{
		t.Fatalf("ошибка в смене статуса: %v", err)
	}
	if order.Status != StatusConfirmed{
		t.Errorf("статус не confirmed: %v", order.Status)
	}
	if !order.UpdatedAt.Equal(changedAt){
		t.Errorf("время апдейта неправильное: %v", order.UpdatedAt)
	}
	if !order.CreatedAt.Equal(createdAt){
		t.Errorf("время криейта не правильное: %v", order.CreatedAt)
	}
}

func TestOrderChangeStatusNotAllowed(t *testing.T){
	createdAt := time.Date(2026, 8, 16, 18, 1, 0, 0, time.UTC)
	changedAt := time.Date(2026, 8, 16, 18, 2, 0, 0, time.UTC)
	order := &Order{Status: StatusCreated, CreatedAt: createdAt, UpdatedAt: createdAt}

	err := order.ChangeStatus(StatusShipped, changedAt)
	if !errors.Is(err, ErrStatusTransition) {
		t.Errorf("ожидали: %v, получили:%v ", ErrStatusTransition, err)
	}
	if order.Status != StatusCreated{
		t.Errorf("должен быть status created: %v", order.Status)
	}
	if !order.UpdatedAt.Equal(createdAt){
		t.Errorf("время апдейта неправильное: %v", order.UpdatedAt)
	}
	if !order.CreatedAt.Equal(createdAt){
		t.Errorf("время криейта не правильное: %v", order.CreatedAt)
	}
}
