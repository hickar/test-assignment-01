package events

type OrderStatus string

const (
	OrderStatusCreated  OrderStatus = "CREATED"
	OrderStatusPaid     OrderStatus = "PAID"
	OrderStatusCanceled OrderStatus = "CANCELED"
)

type OrderCreatedEvent struct {
	ID          int64       `json:"id"`
	OrderID     int64       `json:"order_id"`
	UserID      int64       `json:"user_id"`
	Status      OrderStatus `json:"status"`
	AmountCents int64       `json:"amount_cents"`
}

type AccountOrderPaymentStatus string

const (
	AccountOrderStatusCanceled AccountOrderPaymentStatus = "CANCELED"
	AccountOrderStatusPaid     AccountOrderPaymentStatus = "PAID"
)

type AccountOrderPaymentEvent struct {
	ID           int64                     `json:"id"`
	OrderEventID int64                     `json:"order_event_id"`
	AccountID    int64                     `json:"account_id"`
	OrderID      int64                     `json:"order_id"`
	Status       AccountOrderPaymentStatus `json:"status"`
}
