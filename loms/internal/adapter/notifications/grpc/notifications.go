package grpc

import (
	"context"
	"fmt"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/port"
	notificationsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/notifications/api/v1"
)

type notificationsClient struct {
	client notificationsv1.NotificationsClient
}

func NewNotificationsClient(client notificationsv1.NotificationsClient) *notificationsClient {
	return &notificationsClient{
		client: client,
	}
}

func (c notificationsClient) SendMessage(
	ctx context.Context,
	userID,
	orderID int64,
	status port.OrderStatus,
) error {
	_, err := c.client.SendMessage(ctx, &notificationsv1.SendMessageRequest{
		UserId:  userID,
		OrderId: orderID,
		Status:  FromOrderStatus(status),
	})
	if err != nil {
		return fmt.Errorf(
			"send notification message user_id=%d order_id=%d status=%s: %w",
			userID,
			orderID,
			status,
			err,
		)
	}

	return nil
}
