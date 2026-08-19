package memory

import (
	"context"
	"sync"

	orderdomain "github.com/gee-kul/orderflow/internal/order"
)

type MemoryOrderRepository struct {
	orders map[string]orderdomain.Order
	mu     sync.RWMutex
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	var memoryOrder MemoryOrderRepository
	memoryOrder.orders = make(map[string]orderdomain.Order)

	return &memoryOrder
}

func (r *MemoryOrderRepository) Save(ctx context.Context, order orderdomain.Order) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	copyOrder := order
	copyOrder.Items = make([]orderdomain.OrderItem, len(order.Items))
	copy(copyOrder.Items, order.Items)
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = copyOrder

	return nil
}

func (r *MemoryOrderRepository) FindByID(ctx context.Context, id string) (orderdomain.Order, error) {
	err := ctx.Err()
	if err != nil {
		return orderdomain.Order{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.orders[id]
	if !ok {
		return orderdomain.Order{}, orderdomain.ErrOrderNotFound
	}
	copyV := v
	copyV.Items = make([]orderdomain.OrderItem, len(v.Items))
	copy(copyV.Items, v.Items)
	return copyV, nil
}
