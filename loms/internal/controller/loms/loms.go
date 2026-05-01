package loms

import (
	"context"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	"go.uber.org/zap"
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mocks.go -package=mocks
type (
	//nolint:revive // LomsService name is intentionally explicit because this package has multiple service interfaces.
	LomsService interface {
		CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (orderID int64, err error)
		GetOrder(ctx context.Context, orderID int64) (order entity.Order, err error)
		PayOrder(ctx context.Context, orderID int64) error
		CancelOrder(ctx context.Context, orderID int64) error
	}
)

type lomsServer struct {
	lomsv1.UnimplementedLomsServer
	lomsService LomsService
	logger      *zap.Logger
}

func NewLomsServer(lomsService LomsService, logger *zap.Logger) *lomsServer {
	return &lomsServer{
		lomsService: lomsService,
		logger:      logger,
	}
}



