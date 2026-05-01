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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *lomsServer) GetOrder(ctx context.Context, req *lomsv1.GetOrderRequest) (*lomsv1.GetOrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	order, err := s.lomsService.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, entity.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		s.logger.Error(
			"failed to get order to loms",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &lomsv1.GetOrderResponse{
		UserId:    order.UserID,
		Items:     converter.FromOrderItems(order.Items),
		Status:    converter.FromOrderStatus(order.Status),
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}, nil
}



