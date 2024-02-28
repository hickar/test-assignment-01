package repository

import (
	"context"
	"database/sql"

	"github.com/hickar/crtex_test_assignment/order_service/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(conn *sql.DB) *OrderRepository {
	return &OrderRepository{db: conn}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	query := `INSERT INTO orders (amount_cents, user_id, status) VALUES ($1, $2, $3) RETURNING id;`

	err := r.db.QueryRowContext(
		ctx,
		query,
		order.AmountCents,
		order.UserID,
		domain.OrderStatusCreated,
	).Scan(&order.ID)
	if err != nil {
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

	return order, err
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, order domain.Order) error {
	query := `UPDATE orders SET status = $2 WHERE id = $1;`

	_, err := r.db.ExecContext(ctx, query, order.ID, order.Status)
	return err
}
