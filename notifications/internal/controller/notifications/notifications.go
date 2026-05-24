package notifications

import (
	"context"

	notificationsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/notifications/api/v1"
	"go.uber.org/zap"
)

//go:generate mockgen -source=notifications.go -destination=mocks/notifications_mocks.go -package=mocks
type Service interface {
	SendMessage(ctx context.Context, userID, orderID int64, status string) error
}

type Server struct {
	notificationsv1.UnimplementedNotificationsServer
	notificationsSerice Service
	logger              *zap.Logger
}

func NewNotificationsServer(
	notificationsSerice Service,
	logger *zap.Logger,
) *Server {
	return &Server{
		notificationsSerice: notificationsSerice,
		logger:              logger,
	}
}
