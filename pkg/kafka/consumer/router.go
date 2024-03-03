package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type TopicRouter struct {
	routes map[string]RouteHandler
}

func NewTopicRouter() *TopicRouter {
	return &TopicRouter{
		routes: make(map[string]RouteHandler),
	}
}

func (r *TopicRouter) Handle(topic string, handler RouteHandler, middlewareFns ...RouteMiddleware) {
	h := handler

	for _, middlewareFn := range middlewareFns {
		h = middlewareFn(h)
	}

	r.routes[topic] = h
}

func (r *TopicRouter) Route(ctx context.Context, message *kafka.Message) {
	handler, ok := r.routes[message.Topic]
	if !ok {
		return
	}

	_ = handler(ctx, message)
}
