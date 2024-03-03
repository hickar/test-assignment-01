package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hickar/crtex_test_assignment/events"

	"github.com/hickar/crtex_test_assignment/order/internal/domain"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: pool}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return order, err
	}
	//nolint:errcheck
	defer tx.Rollback(ctx)

	query := `INSERT INTO orders (amount_cents, user_id, status) VALUES ($1, $2, $3) RETURNING id;`

	err = tx.QueryRow(
		ctx,
		query,
		order.AmountCents,
		order.UserID,
		events.OrderStatusCreated,
	).Scan(&order.ID)
	if err != nil {
		return order, err
	}

	query = `INSERT INTO order_create_events (order_id, amount_cents, user_id) VALUES ($1, $2, $3);`
	_, err = tx.Exec(
		ctx,
		query,
		order.ID,
		order.AmountCents,
		order.UserID,
	)
	if err != nil {
		return order, err
	}

	if err = tx.Commit(ctx); err != nil {
		return order, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int64) (domain.Order, error) {
	query := `SELECT id, user_id, amount_cents, status FROM orders WHERE id = $1;`

	var order domain.Order
	err := r.db.QueryRow(
		ctx,
		query,
		orderID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.AmountCents,
		&order.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return order, domain.ErrNotFound
	}

	return order, err
}

func (r *OrderRepository) UpdateOrderStatusByID(ctx context.Context, orderID int64, orderStatus string) error {
	query := `UPDATE orders SET status = $2 WHERE id = $1;`

	_, err := r.db.Exec(ctx, query, orderID, orderStatus)
	return err
}
