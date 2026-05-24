package port

type OrderStatusChangedNotification struct {
	OrderID int64       `json:"order_id"`
	UserID  int64       `json:"user_id"`
	Status  OrderStatus `json:"status"`
}

type OrderStatus string

const (
	OrderStatusUnspecified     OrderStatus = "unspecified"
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
)
