package cart

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *cartServer) AddItem(ctx context.Context, req *cartv1.AddItemRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.itemService.AddItem(ctx, req.GetUserId(), req.GetSku(), req.GetCount()); err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		case errors.Is(err, entity.ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, "insufficient stock")
		default:
			s.logger.Error(
				"failed to add item to cart",
				zap.Error(err),
			)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}
