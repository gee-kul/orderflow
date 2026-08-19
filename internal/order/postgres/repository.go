package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/gee-kul/orderflow/internal/event"
	orderdomain "github.com/gee-kul/orderflow/internal/order"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	repo := PostgresOrderRepository{}
	repo.pool = pool
	return &repo
}

func saveOrderInTx(ctx context.Context, tx pgx.Tx, order orderdomain.Order) error {

	query := `INSERT INTO orders(id, customer_id, status, total_amount, currency, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id)
	DO UPDATE SET customer_id = EXCLUDED.customer_id, status = EXCLUDED.status,
	total_amount = EXCLUDED.total_amount, currency = EXCLUDED.currency,
	created_at = EXCLUDED.created_at, updated_at = EXCLUDED.updated_at`
	_, err := tx.Exec(ctx, query, order.ID, order.CustomerID, order.Status, order.TotalAmount,
		order.Currency, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("не удалось сохранить заказ: %w", err)
	}

	query2 := `DELETE FROM order_items WHERE order_id = $1`
	_, err = tx.Exec(ctx, query2, order.ID)
	if err != nil {
		return fmt.Errorf("не удалось удалить старые позиции заказа: %w", err)
	}

	for pos, item := range order.Items {
		query3 := `INSERT INTO order_items(order_id, position,
		product_id, name, unit_price, quantity)
		VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := tx.Exec(ctx, query3, order.ID, pos, item.ProductID, item.Name, item.UnitPrice, item.Quantity)
		if err != nil {
			return fmt.Errorf("ошибка на %d позиции: %w", pos, err)
		}
	}
	return nil
}

func (p *PostgresOrderRepository) Save(ctx context.Context, order orderdomain.Order) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка открытия транзакции :%w", err)
	}
	defer tx.Rollback(ctx)

	err = saveOrderInTx(ctx, tx, order)
	if err != nil {
		return fmt.Errorf("не удалось сохранить заказ в транзакции %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("ошибка комита: %w", err)
	}
	return nil
}

func (p *PostgresOrderRepository) SaveWithEvent(ctx context.Context, order orderdomain.Order, evt event.Event) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка открытия транзакции :%w", err)
	}
	defer tx.Rollback(ctx)

	err = saveOrderInTx(ctx, tx, order)
	if err != nil {
		return fmt.Errorf("не удалось сохранить заказ в транзакции %w", err)
	}

	query := `INSERT INTO outbox_events(id, aggregate_type, aggregate_id, event_type, payload, created_at, published_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = tx.Exec(ctx, query, evt.ID, evt.AggregateType, evt.AggregateID, evt.EventType,
		evt.Payload, evt.CreatedAt, evt.PublishedAt)
	if err != nil {
		return fmt.Errorf("не удалось сохранить outbox-событие: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("ошибка комита: %w", err)
	}
	return nil
}

func (p *PostgresOrderRepository) FindByID(ctx context.Context, id string) (orderdomain.Order, error) {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("ошибка открытия транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	order := orderdomain.Order{}

	query := `SELECT id, customer_id, status, total_amount,
	currency, created_at, updated_at FROM orders WHERE id = $1`
	row := tx.QueryRow(ctx, query, id)

	err = row.Scan(&order.ID, &order.CustomerID, &order.Status, &order.TotalAmount,
		&order.Currency, &order.CreatedAt, &order.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return orderdomain.Order{}, orderdomain.ErrOrderNotFound
	}
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("ошибка при сканировании столбцов: %w", err)
	}
	query2 := `SELECT product_id, name, unit_price, quantity FROM order_items
	WHERE order_id = $1 ORDER BY position ASC`
	rows, err := tx.Query(ctx, query2, order.ID)
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("не удалось получить позиции заказа: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item := orderdomain.OrderItem{}
		err = rows.Scan(&item.ProductID, &item.Name, &item.UnitPrice, &item.Quantity)
		if err != nil {
			return orderdomain.Order{}, fmt.Errorf("не удалось получить позиции заказа:%w", err)
		}
		order.Items = append(order.Items, item)
	}
	err = rows.Err()
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("ошибка чтения позиций заказа:%w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return orderdomain.Order{}, fmt.Errorf("ошибка комита: %w", err)
	}
	return order, nil
}
