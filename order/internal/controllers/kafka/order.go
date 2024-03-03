package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"

	"github.com/hickar/crtex_test_assignment/events"
	"github.com/hickar/crtex_test_assignment/order/internal/domain"
)

type OrderHandler struct {
	service domain.Service
}

func NewOrderHandler(service domain.Service) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) NewAccountOrderEvent(ctx context.Context, message *kafka.Message) error {
	var eventMsg AccountMessage
	if err := json.Unmarshal(message.Value, &eventMsg); err != nil {
		return err
	}

	err := h.service.UpdateOrder(ctx, eventMsg.Payload)
	return err
}

type AccountMessage struct {
	Payload events.AccountOrderPaymentEvent `json:"payload"`
}
