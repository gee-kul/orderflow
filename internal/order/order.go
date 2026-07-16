package order

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

func NewOrder(id string, customerID string, currency string,
	items []OrderItem, now time.Time) (*Order, error) {

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrOrderIDRequired
	}

	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerIDRequired
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return nil, ErrCurrencyRequired
	}

	if len(items) == 0 {
		return nil, ErrItemsRequired
	}

	var finalPrice int64
	itemsCopy := slices.Clone(items)
	for i := range itemsCopy {
		item := &itemsCopy[i]
		item.ProductID = strings.TrimSpace(item.ProductID)
		if item.ProductID == "" {
			return nil, fmt.Errorf("item %d: %w", i+1, ErrProductIDRequired)
		}

		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			return nil, fmt.Errorf("item %d: %w", i+1, ErrProductNameRequired)
		}

		if item.UnitPrice <= 0 {
			return nil, fmt.Errorf("item %d: %w", i+1, ErrUnitPriceInvalid)
		}

		if item.Quantity <= 0 {
			return nil, fmt.Errorf("item %d: %w", i+1, ErrQuantityInvalid)
		}

		if item.UnitPrice > math.MaxInt64/int64(item.Quantity) {
			return nil, fmt.Errorf("item %d: %w", i+1, ErrTotalAmountOverflow)
		}

		priceForPosition := item.UnitPrice * int64(item.Quantity)

		if finalPrice > math.MaxInt64-priceForPosition {
			return nil, fmt.Errorf("item %d: %w", i+1, ErrTotalAmountOverflow)
		}
		finalPrice += priceForPosition
	}

	oneOrder := &Order{ID: id, CustomerID: customerID, Items: itemsCopy,
		Status: StatusCreated, TotalAmount: finalPrice, Currency: currency,
		CreatedAt: now, UpdatedAt: now}

	return oneOrder, nil
}
