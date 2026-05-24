package converter

import (
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
)

func ToOrderStatus(status lomsv1.OrderStatus) entity.OrderStatus {
	switch status {
	case lomsv1.OrderStatus_ORDER_STATUS_NEW:
		return entity.OrderStatusNew
	case lomsv1.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT:
		return entity.OrderStatusAwaitingPayment
	case lomsv1.OrderStatus_ORDER_STATUS_PAID:
		return entity.OrderStatusPaid
	case lomsv1.OrderStatus_ORDER_STATUS_FAILED:
		return entity.OrderStatusFailed
	case lomsv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return entity.OrderStatusCancelled
	case lomsv1.OrderStatus_ORDER_STATUS_UNSPECIFIED:
		return entity.OrderStatusUnknown
	default:
		return entity.OrderStatusUnknown
	}
}

func FromOrderStatus(status entity.OrderStatus) lomsv1.OrderStatus {
	switch status {
	case entity.OrderStatusNew:
		return lomsv1.OrderStatus_ORDER_STATUS_NEW
	case entity.OrderStatusAwaitingPayment:
		return lomsv1.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case entity.OrderStatusPaid:
		return lomsv1.OrderStatus_ORDER_STATUS_PAID
	case entity.OrderStatusFailed:
		return lomsv1.OrderStatus_ORDER_STATUS_FAILED
	case entity.OrderStatusCancelled:
		return lomsv1.OrderStatus_ORDER_STATUS_CANCELLED
	case entity.OrderStatusUnknown:
		return lomsv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	default:
		return lomsv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func FromOrderItems(items []entity.OrderItem) []*lomsv1.Item {
	lomItems := make([]*lomsv1.Item, len(items))
	for i, item := range items {
		lomItems[i] = &lomsv1.Item{
			Sku:   item.SKU,
			Count: uint32(item.Count),
		}
	}
	return lomItems
}

func ToOrderItems(items []*lomsv1.Item) []entity.OrderItem {
	orderItems := make([]entity.OrderItem, len(items))
	for i, item := range items {
		orderItems[i] = entity.OrderItem{
			SKU:   item.GetSku(),
			Count: uint64(item.GetCount()),
		}
	}
	return orderItems
}
