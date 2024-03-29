package domain

import (
	"context"
	"fmt"

	"github.com/hickar/crtex_test_assignment/events"
)

type Service interface {
	CreateOrder(context.Context, Order) (int64, error)
	GetOrderByID(context.Context, int64) (Order, error)
	UpdateOrder(context.Context, events.AccountOrderPaymentEvent) error
}

type Order struct {
	ID          int64
	UserID      int64
	AmountCents int64
	Status      string
}

type OrderRepository interface {
	GetOrderByID(context.Context, int64) (Order, error)
	CreateOrder(context.Context, Order) (Order, error)
	UpdateOrderStatusByID(context.Context, int64, string) error
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

	if !isValidOrderPayload(order) {
		return orderID, ErrInvalidData
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

func (s *OrderService) UpdateOrder(ctx context.Context, event events.AccountOrderPaymentEvent) error {
	if !isValidOrderEventPayload(event) {
		return ErrInvalidData
	}

	err := s.repo.UpdateOrderStatusByID(
		ctx,
		event.OrderID,
		string(event.Status),
	)
	return err
}

func isValidOrderPayload(order Order) bool {
	return order.AmountCents > 0
}

func isValidOrderEventPayload(event events.AccountOrderPaymentEvent) bool {
	if event.Status != events.AccountOrderStatusCanceled && event.Status != events.AccountOrderStatusPaid {
		return false
	}

	return true
}
