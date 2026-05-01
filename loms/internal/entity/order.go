package entity

import (
	"errors"
	"time"
)

type Order struct {
	ID        int64
	UserID    int64
	Items     []OrderItem
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	SKU   uint32
	Count uint64
}
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusUnknown         OrderStatus = "unknown"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidStatus      = errors.New("invalid order status for operation")
	ErrOrderAlreadyPaid   = errors.New("cannot cancel a paid order")
	ErrOrderCancelled     = errors.New("order is cancelled")
	ErrOrderFailed        = errors.New("order is failed")
	ErrUnknownOrderStatus = errors.New("unknown order status")
)



