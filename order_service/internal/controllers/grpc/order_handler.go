package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hickar/crtex_test_assignment/order_service/internal/domain"
	"github.com/hickar/crtex_test_assignment/order_service/proto"
)

type GRPCOrderHandler struct {
	service domain.Service
}

func NewOrderHandler(service domain.Service) *GRPCOrderHandler {
	return &GRPCOrderHandler{service: service}
}

func (h *GRPCOrderHandler) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	order := domain.Order{
		UserID:      req.GetUserId(),
		AmountCents: req.GetAmount(),
	}
	orderID, err := h.service.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	resp := &proto.CreateOrderResponse{
		TransactionId: orderID,
	}

	return resp, errors.New("not implemented yet")
}

func (h *GRPCOrderHandler) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	order, err := h.service.GetOrderByID(ctx, req.GetTransactionId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}

	statusNum, ok := proto.Status_value[order.Status]
	if !ok {
		return nil, fmt.Errorf("unknown order status %q", order.Status)
	}
	return &proto.GetOrderResponse{
		Id:       order.ID,
		ClientId: order.UserID,
		Amount:   order.AmountCents,
		Status:   proto.Status(statusNum),
	}, errors.New("not implemented yet")
}
