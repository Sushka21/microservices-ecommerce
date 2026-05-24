package loms

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *lomsServer) CancelOrder(ctx context.Context, req *lomsv1.CancelOrderRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.lomsService.CancelOrder(ctx, req.GetOrderId()); err != nil {
		switch {
		case errors.Is(err, entity.ErrOrderCancelled),
			errors.Is(err, entity.ErrOrderAlreadyPaid):
			return &emptypb.Empty{}, status.Error(codes.FailedPrecondition, "err cancel order")
		default:
			s.logger.Error(
				"failed to cancel order to loms",
				zap.Error(err),
			)
			return &emptypb.Empty{}, status.Error(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}
