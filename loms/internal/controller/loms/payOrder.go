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

func (s *lomsServer) PayOrder(ctx context.Context, req *lomsv1.PayOrderRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.lomsService.PayOrder(ctx, req.GetOrderId()); err != nil {
		switch {
		case errors.Is(err, entity.ErrOrderNotFound):
			return &emptypb.Empty{}, status.Error(codes.NotFound, "order nor found")
		case errors.Is(err, entity.ErrOrderAlreadyPaid),
			errors.Is(err, entity.ErrOrderCancelled),
			errors.Is(err, entity.ErrOrderFailed),
			errors.Is(err, entity.ErrUnknownOrderStatus):
			return &emptypb.Empty{}, status.Error(codes.FailedPrecondition, "error pay order")
		default:
			s.logger.Error(
				"failed to pay order to loms",
				zap.Error(err),
			)
			return &emptypb.Empty{}, status.Error(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}



