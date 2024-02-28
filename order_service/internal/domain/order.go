package domain

import (
	"context"
	"fmt"
)

type Service interface {
	CreateOrder(context.Context, Order) (int64, error)
	GetOrderByID(context.Context, int64) (Order, error)
	UpdateOrder(context.Context, Order) error
}

type Order struct {
	ID          int64
	UserID      int64
	AmountCents int64
	Status      string
}

const (
	OrderStatusCreated   = "created"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"
)

type OrderEvent struct {
	ID      int64
	OrderID int64
}

type OrderRepository interface {
	GetOrderByID(context.Context, int64) (Order, error)
	CreateOrder(context.Context, Order) (Order, error)
	UpdateOrder(context.Context, Order) error
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, order Order) (int64, error) {
	var orderID int64
	var err error

	if order.AmountCents < 0 {
		return orderID, nil
	}

	order, err = s.repo.CreateOrder(ctx, order)
	if err != nil {
		return orderID, fmt.Errorf("failed to create order: %w", err)
	}

	orderID = order.ID
	return orderID, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID int64) (Order, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return order, fmt.Errorf("failed to retrieve order by id: %w", err)
	}

	return order, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, order Order) error {
	// TODO: implement
	return nil
}
