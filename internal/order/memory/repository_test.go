package memory

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	orderdomain "github.com/gee-kul/orderflow/internal/order"
)

func TestMemoryOrderRepositorySaveAndFindByID(t *testing.T) {
	repository := NewMemoryOrderRepository()

	order := orderdomain.Order{ID: "order-1"}

	err := repository.Save(context.Background(), order)
	if err != nil {
		t.Fatalf("save не должен был вернуть ошибку: %v", err)
	}
	ord, err := repository.FindByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("find не должен был вернуть ошибку: %v", err)
	}
	if ord.ID != order.ID {
		t.Errorf("айди должны были совпать: %v", ord.ID)
	}
}

func TestMemoryOrderRepositoryFindByIDNotFound(t *testing.T) {
	repository := NewMemoryOrderRepository()

	order, err := repository.FindByID(context.Background(), "unknown-order")
	if !errors.Is(err, orderdomain.ErrOrderNotFound) {
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
			order := orderdomain.Order{ID: "order-" + strconv.Itoa(i)}
			err := repository.Save(context.Background(), order)
			if err != nil {
				t.Errorf("не удалось сохранить заказ %s:%v", order.ID, err)
			}
		}(i)
	}
	wg.Wait()
	for i := range 10 {
		id := "order-" + strconv.Itoa(i)
		ord, err := repository.FindByID(context.Background(), id)
		if err != nil {
			t.Errorf("не удалось найти заказ %s:%v", id, err)
		}
		if ord.ID != id {
			t.Errorf("айди не совпадают %v:%v", ord.ID, id)
		}
	}
}

func TestSaveWithCancelledContext(t *testing.T) {
	repository := NewMemoryOrderRepository()

	order := orderdomain.Order{ID: "order-1"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err2 := repository.Save(ctx, order)
	if !errors.Is(err2, context.Canceled) {
		t.Errorf("ожидаемая и полученная ошибки не совпали: %v", err2)
	}
	_, err := repository.FindByID(context.Background(), order.ID)
	if !errors.Is(err, orderdomain.ErrOrderNotFound) {
		t.Errorf("отмененная операция сохранилась %v", err)
	}
}

func TestFindByIDWithCancelledContext(t *testing.T) {
	repository := NewMemoryOrderRepository()
	order := orderdomain.Order{ID: "order-1"}
	err := repository.Save(context.Background(), order)
	if err != nil {
		t.Fatalf("save не должен был вернуть ошибку: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = repository.FindByID(ctx, order.ID)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ожидалась ошибка отмены контекста %v", err)
	}
}

func TestDefensiveCopyingForSave(t *testing.T) {
	repository := NewMemoryOrderRepository()
	item := orderdomain.OrderItem{ProductID: "product-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	order := orderdomain.Order{ID: "order-1", Items: []orderdomain.OrderItem{item}}

	err := repository.Save(context.Background(), order)
	if err != nil {
		t.Fatalf("ошибка сохранения заказа: %v", err)
	}

	order.Items[0].Name = "not bruh"

	ord, err := repository.FindByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("ошибка поиска по айди %v", err)
	}
	if ord.Items[0].Name != "bruh" {
		t.Errorf("имя товара изменилось но не должно было %v", ord.Items[0].Name)
	}
}

func TestDefensiveCopyingForFindByID(t *testing.T) {
	repository := NewMemoryOrderRepository()
	item := orderdomain.OrderItem{ProductID: "product-1", Name: "bruh", UnitPrice: 100500, Quantity: 1}
	order := orderdomain.Order{ID: "order-1", Items: []orderdomain.OrderItem{item}}

	err := repository.Save(context.Background(), order)
	if err != nil {
		t.Fatalf("ошибка сохранения заказа: %v", err)
	}

	ord, err := repository.FindByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("ошибка поиска по айди %v", err)
	}

	ord.Items[0].Name = "not bruh"

	ord2, err := repository.FindByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("ошибка поиска по айди %v", err)
	}
	if ord2.Items[0].Name != "bruh" {
		t.Errorf("имя товара изменилось но не должно было %v", ord2.Items[0].Name)
	}
}
