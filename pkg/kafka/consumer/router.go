package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type ConsumerTopicRouter struct {
	routes map[string]RouteHandler
}

func NewConsumerTopicRouter() *ConsumerTopicRouter {
	return &ConsumerTopicRouter{
		routes: make(map[string]RouteHandler),
	}
}

func (r *ConsumerTopicRouter) Handle(topic string, handler RouteHandler, middlewareFns ...RouteMiddleware) {
	h := handler

	for _, middlewareFn := range middlewareFns {
		h = middlewareFn(h)
	}

	r.routes[topic] = h
}

func (r *ConsumerTopicRouter) Route(ctx context.Context, message *kafka.Message) {
	handler, ok := r.routes[message.Topic]
	if !ok {
		return
	}

	handler(ctx, message)
}
