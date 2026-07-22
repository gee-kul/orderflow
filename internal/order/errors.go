package order

import "errors"

var (
	ErrOrderIDRequired     = errors.New("order ID is required")
	ErrCustomerIDRequired  = errors.New("customer ID is required")
	ErrItemsRequired       = errors.New("order item is required")
	ErrCurrencyRequired    = errors.New("currency is required")
	ErrProductIDRequired   = errors.New("product ID is required")
	ErrProductNameRequired = errors.New("product NAME is required")
	ErrUnitPriceInvalid    = errors.New("unit price is invalid")
	ErrQuantityInvalid     = errors.New("quantity is invalid")
	ErrTotalAmountOverflow = errors.New("total amount overflow")
	ErrStatusTransition    = errors.New("status transition is not allowed")
)
