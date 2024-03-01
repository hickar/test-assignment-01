package domain

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrder(t *testing.T) {
	repo := newOrderRepoStub(nil, func(_ context.Context, order Order) (Order, error) {
		//nolint:gosec
		order.ID = rand.Int63() + 1
		return order, nil
	}, nil)
	service := NewOrderService(repo)

	tests := []struct {
		name      string
		order     Order
		shouldErr bool
		err       error
	}{
		{
			name: "ValidOrder",
			order: Order{
				UserID:      1,
				AmountCents: 10000,
			},
		},
		{
			name: "Invalid_NegativeAmount",
			order: Order{
				AmountCents: -1,
			},
			shouldErr: true,
			err:       ErrInvalidData,
		},
		{
			name:      "Invlalid_ZeroAmount",
			order:     Order{},
			shouldErr: true,
			err:       ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderID, err := service.CreateOrder(context.Background(), tt.order)
			if tt.shouldErr {
				assert.ErrorIs(t, err, tt.err)
				assert.Zero(t, orderID)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, orderID)
		})
	}
}

func TestGetOrderByID(t *testing.T) {
	repo := newOrderRepoStub(func(_ context.Context, orderID int64) (Order, error) {
		if orderID == 0 {
			return Order{}, ErrNotFound
		}

		return Order{
			ID: orderID,
			//nolint:gosec
			UserID: rand.Int63() + 1,
			//nolint:gosec
			AmountCents: rand.Int63() + 1,
		}, nil
	}, nil, nil)
	service := NewOrderService(repo)

	tests := []struct {
		name      string
		orderID   int64
		shouldErr bool
		err       error
	}{
		{
			name:    "ValidOrder",
			orderID: 1,
		},
		{
			name:      "Invalid_NotFound",
			orderID:   0,
			shouldErr: true,
			err:       ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := service.GetOrderByID(context.Background(), tt.orderID)
			if tt.shouldErr {
				assert.ErrorIs(t, err, tt.err)
				assert.Zero(t, order.ID)
				return
			}

			assert.NoError(t, err, "no error must occur, got %s", err)
			assert.Equal(t, tt.orderID, order.ID)
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	repo := newOrderRepoStub(nil, nil, func(_ context.Context, orderID int64, orderStatus string) error {
		return nil
	})
	service := NewOrderService(repo)

	tests := []struct {
		name       string
		orderEvent OrderEvent
		shouldErr  bool
		err        error
	}{
		{
			name: "Valid",
			orderEvent: OrderEvent{
				Status: OrderStatusPaid,
			},
		},
		{
			name: "Invalid_NotFound",
			orderEvent: OrderEvent{
				Status: "NonExistentStatus",
			},
			shouldErr: true,
			err:       ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateOrder(context.Background(), tt.orderEvent)
			if tt.shouldErr {
				assert.ErrorIs(t, err, tt.err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

type orderRepoStub struct {
	getOrderByIDFn      func(context.Context, int64) (Order, error)
	createOrderFn       func(context.Context, Order) (Order, error)
	updateOrderStatusFn func(context.Context, int64, string) error
}

func (r *orderRepoStub) GetOrderByID(ctx context.Context, orderID int64) (Order, error) {
	return r.getOrderByIDFn(ctx, orderID)
}

func (r *orderRepoStub) CreateOrder(ctx context.Context, order Order) (Order, error) {
	return r.createOrderFn(ctx, order)
}

func (r *orderRepoStub) UpdateOrderStatusByID(ctx context.Context, orderID int64, orderStatus string) error {
	return r.updateOrderStatusFn(ctx, orderID, orderStatus)
}

func newOrderRepoStub(
	getOrderByIDFn func(context.Context, int64) (Order, error),
	createOrderFn func(context.Context, Order) (Order, error),
	updateOrderStatusFn func(context.Context, int64, string) error,
) *orderRepoStub {
	return &orderRepoStub{
		getOrderByIDFn:      getOrderByIDFn,
		createOrderFn:       createOrderFn,
		updateOrderStatusFn: updateOrderStatusFn,
	}
}
