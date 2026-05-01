package controller

import (
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/controller/notifications"
	notificationsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/notifications/api/v1"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type API struct {
	NotificationsServer notificationsv1.NotificationsServer
	logger              *zap.Logger
}

func New(notificationsSetvice notifications.Service, logger *zap.Logger) *API {
	return &API{
		NotificationsServer: notifications.NewNotificationsServer(notificationsSetvice, logger),
		logger:              logger,
	}
}

func (a *API) Register(grpcServer *grpc.Server) {
	notificationsv1.RegisterNotificationsServer(grpcServer, a.NotificationsServer)
}



