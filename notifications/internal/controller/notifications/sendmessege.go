package notifications

import (
	"context"
	"errors"

	usecase "github.com/Sushka21/microservices-ecommerce/notifications/internal/usecase/notifications"
	notificationsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/notifications/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) SendMessage(ctx context.Context, req *notificationsv1.SendMessageRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.notificationsSerice.SendMessage(ctx, req.GetUserId(), req.GetOrderId(), req.GetStatus().String()); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, "error context canceld")
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "error DeadlineExceeded ")
		case errors.Is(err, usecase.ErrCallbackRequestFailed):
			return nil, status.Error(codes.Unavailable, "error CallbackRequestFailed")
		case errors.Is(err, usecase.ErrCallbackBadStatus):
			return nil, status.Error(codes.Unavailable, "error bad status")
		default:
			s.logger.Error(
				"failed to send message to notifications",
				zap.Error(err),
			)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}
