package order

import (
	"context"
	"sync"
)

type MemoryOrderRepository struct {
	orders map[string]Order
	mu     sync.RWMutex
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	var memoryOrder MemoryOrderRepository
	memoryOrder.orders = make(map[string]Order)

	return &memoryOrder
}

func (r *MemoryOrderRepository) Save(ctx context.Context, order Order) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	copyOrder := order
	copyOrder.Items = make([]OrderItem, len(order.Items))
	copy(copyOrder.Items, order.Items)
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = copyOrder

	return nil
}

func (r *MemoryOrderRepository) FindByID(ctx context.Context, id string) (Order, error) {
	err := ctx.Err()
	if err != nil {
		return Order{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.orders[id]
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	copyV := v
	copyV.Items = make([]OrderItem, len(v.Items))
	copy(copyV.Items, v.Items)
	return copyV, nil
}
