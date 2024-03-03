package kafka

import (
	"context"
	"encoding/json"

	"github.com/hickar/crtex_test_assignment/events"

	"github.com/segmentio/kafka-go"

	"github.com/hickar/crtex_test_assignment/account/internal/domain"
)

type AccountHandler struct {
	service domain.Service
}

func NewAccountHandler(service domain.Service) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) NewOrderEvent(ctx context.Context, message *kafka.Message) error {
	var eventMsg OrderCreatedMessage
	if err := json.Unmarshal(message.Value, &eventMsg); err != nil {
		return err
	}

	err := h.service.ProcessNewOrder(ctx, eventMsg.Payload)
	return err
}

type OrderCreatedMessage struct {
	Payload events.OrderCreatedEvent `json:"payload"`
}
