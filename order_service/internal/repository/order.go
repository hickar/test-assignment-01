package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hickar/crtex_test_assignment/order_service/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(conn *sql.DB) *OrderRepository {
	return &OrderRepository{db: conn}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return order, err
	}
	defer tx.Rollback()

	query := `INSERT INTO orders (amount_cents, user_id, status) VALUES ($1, $2, $3) RETURNING id;`

	err = tx.QueryRowContext(
		ctx,
		query,
		order.AmountCents,
		order.UserID,
		domain.OrderStatusCreated,
	).Scan(&order.ID)
	if err != nil {
		return order, err
	}

	query = `INSERT INTO order_create_events (order_id, amount_cents, user_id) VALUES ($1, $2, $3);`
	_, err = tx.ExecContext(
		ctx,
		query,
		order.ID,
		order.AmountCents,
		order.UserID,
	)
	if err != nil {
		return order, err
	}

	if err = tx.Commit(); err != nil {
		return order, err
	}

	return order, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID int64) (domain.Order, error) {
	query := `SELECT id, user_id, amount_cents, status FROM orders WHERE id = $1;`

	var order domain.Order
	err := r.db.QueryRowContext(
		ctx,
		query,
		orderID,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.AmountCents,
		&order.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return order, domain.ErrNotFound
	}

	return order, err
}

func (r *OrderRepository) UpdateOrderStatusByID(ctx context.Context, orderID int64, orderStatus string) error {
	query := `UPDATE orders SET status = $2 WHERE id = $1;`

	_, err := r.db.ExecContext(ctx, query, orderID, orderStatus)
	return err
}
