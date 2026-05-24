package loms

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/converter"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *lomsServer) CreateOrder(ctx context.Context, req *lomsv1.CreateOrderRequest) (*lomsv1.CreateOrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	orderID, err := s.lomsService.CreateOrder(ctx, req.GetUserId(), converter.ToOrderItems(req.GetItems()))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, "InsufficientStock")
		default:
			s.logger.Error(
				"failed to creat order to loms",
				zap.Error(err),
			)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &lomsv1.CreateOrderResponse{
		OrderId: orderID,
	}, nil
}
