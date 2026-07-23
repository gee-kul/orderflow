package order

import "sync"

type MemoryOrderRepository struct {
	orders map[string]Order
	mu     sync.RWMutex
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	var memoryOrder MemoryOrderRepository
	memoryOrder.orders = make(map[string]Order)

	return &memoryOrder
}

func (r *MemoryOrderRepository) Save(order Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order

	return nil
}

func (r *MemoryOrderRepository) FindByID(id string) (Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.orders[id]
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	return v, nil
}
