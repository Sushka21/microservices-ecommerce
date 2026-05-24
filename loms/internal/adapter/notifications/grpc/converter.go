package grpc

import (
	"github.com/Sushka21/microservices-ecommerce/loms/internal/port"
	notificationsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/notifications/api/v1"
)

func FromOrderStatus(orderStatus port.OrderStatus) notificationsv1.OrderStatus {
	switch orderStatus {
	case port.OrderStatusNew:
		return notificationsv1.OrderStatus_ORDER_STATUS_NEW
	case port.OrderStatusAwaitingPayment:
		return notificationsv1.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case port.OrderStatusFailed:
		return notificationsv1.OrderStatus_ORDER_STATUS_FAILED
	case port.OrderStatusPaid:
		return notificationsv1.OrderStatus_ORDER_STATUS_PAID
	case port.OrderStatusCancelled:
		return notificationsv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return notificationsv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
