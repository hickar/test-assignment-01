package grpc

import (
	"context"
	"errors"

	"github.com/hickar/crtex_test_assignment/order/internal/domain"
	"github.com/hickar/crtex_test_assignment/order/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCOrderHandler struct {
	proto.UnimplementedOrderServer
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
		return nil, status.Error(getGRPCErrorCode(err), err.Error())
	}

	resp := &proto.CreateOrderResponse{
		TransactionId: orderID,
	}

	return resp, nil
}

func (h *GRPCOrderHandler) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	order, err := h.service.GetOrderByID(ctx, req.GetTransactionId())
	if err != nil {
		return nil, status.Error(getGRPCErrorCode(err), err.Error())
	}

	statusNum, ok := proto.Status_value[order.Status]
	if !ok {
		return nil, status.Error(codes.Internal, "invalid output status value")
	}

	return &proto.GetOrderResponse{
		Id:       order.ID,
		ClientId: order.UserID,
		Amount:   order.AmountCents,
		Status:   proto.Status(statusNum),
	}, nil
}

func getGRPCErrorCode(err error) codes.Code {
	switch {
	case errors.Is(err, context.Canceled):
		return codes.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return codes.DeadlineExceeded
	case errors.Is(err, domain.ErrNotFound):
		return codes.NotFound
	case errors.Is(err, domain.ErrInvalidData):
		return codes.InvalidArgument
	default:
		return codes.Unknown
	}
}
