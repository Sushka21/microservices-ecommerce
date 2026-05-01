package cart

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *cartServer) CheckoutCart(ctx context.Context, req *cartv1.CheckoutCartRequest) (*cartv1.CheckoutCartResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	orderID, err := s.cartService.CheckoutCart(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrCartIsEmpty):
			return nil, status.Error(codes.FailedPrecondition, "empty cart")
		default:
			s.logger.Error(
				"failed to CheckoutCart to cart",
				zap.Error(err),
			)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &cartv1.CheckoutCartResponse{
		OrderId: orderID,
	}, nil
}



