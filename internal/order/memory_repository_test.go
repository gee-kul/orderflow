package order

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

func TestMemoryOrderRepositorySaveAndFindByID(t *testing.T) {
	repository := NewMemoryOrderRepository()

	order := Order{ID: "order-1"}

	err := repository.Save(order)
	if err != nil {
		t.Fatalf("save не должен был вернуть ошибку: %v", err)
	}
	ord, err := repository.FindByID(order.ID)
	if err != nil {
		t.Fatalf("find не должен был вернуть ошибку: %v", err)
	}
	if ord.ID != order.ID {
		t.Errorf("айди должны были совпать: %v", ord.ID)
	}
}

func TestMemoryOrderRepositoryFindByIDNotFound(t *testing.T) {
	repository := NewMemoryOrderRepository()

	order, err := repository.FindByID("unknown-order")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("ошибки должны были совпасть: %v", err)
	}
	if order.ID != "" {
		t.Errorf("id должен был быть пустым: %v", order.ID)
	}
}

func TestMemoryOrderRepositoryConcurrentSave(t *testing.T) {
	repository := NewMemoryOrderRepository()

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			order := Order{ID: "order-" + strconv.Itoa(i)}
			err := repository.Save(order)
			if err != nil {
				t.Errorf("не удалось сохранить заказ %s:%v", order.ID, err)
			}
		}(i)
	}
	wg.Wait()
	for i := range 10 {
		id := "order-" + strconv.Itoa(i)
		ord, err := repository.FindByID(id)
		if err != nil {
			t.Errorf("не удалось найти заказ %s:%v", id, err)
		}
		if ord.ID != id {
			t.Errorf("айди не совпадают %v:%v", ord.ID, id)
		}
	}
}
