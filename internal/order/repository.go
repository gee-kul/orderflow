package order

import "context"

type OrderRepository interface {
	Save(ctx context.Context, order Order) error
	FindByID(ctx context.Context, id string) (Order, error)
}
