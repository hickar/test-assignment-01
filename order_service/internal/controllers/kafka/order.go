package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/hickar/crtex_test_assignment/order_service/internal/domain"
)

type KafkaOrderHandler struct {
	service domain.Service
}

func NewKafkaOrderHandler(service domain.Service) *KafkaOrderHandler {
	return &KafkaOrderHandler{service: service}
}

func (h *KafkaOrderHandler) NewAccountOrderEvent(ctx context.Context, message *kafka.Message) {
	var accountOrderEvent domain.OrderEvent
	_ = h.service.UpdateOrder(ctx, accountOrderEvent)
}
