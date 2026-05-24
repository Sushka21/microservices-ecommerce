package cart

import (
	"context"

	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *cartServer) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.cartService.ClearCart(ctx, req.GetUserId()); err != nil {
		s.logger.Error(
			"failed to clear cart to cart",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &emptypb.Empty{}, nil
}
