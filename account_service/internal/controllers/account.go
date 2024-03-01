package controllers

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/hickar/crtex_test_assignment/account_service/internal/domain"
)

type AccountHandler struct {
	service domain.Service
}

func NewAccountKafkaHandler(service domain.Service) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) NewOrderEvent(ctx context.Context, message *kafka.Message) {
	var orderEvent domain.IncomingOrderEvent
	_ = h.service.ProcessOrder(ctx, orderEvent)
}
