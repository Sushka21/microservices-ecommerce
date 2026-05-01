package cart

import (
	// "context"

	"errors"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *cartServer) ListCart(req *cartv1.ListCartRequest, stream cartv1.Cart_ListCartServer) error {
	if err := req.Validate(); err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	items, totalPrice, err := s.cartService.ListCart(stream.Context(), req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrCartIsEmpty):
			return nil
		default:
			s.logger.Error(
				"failed to listCart to cart",
				zap.Error(err),
			)
			return status.Error(codes.Internal, "internal error")
		}
	}

	resp := &cartv1.ListCartResponse{
		Items:      make([]*cartv1.Item, len(items)),
		TotalPrice: totalPrice,
	}

	for i, item := range items {
		resp.Items[i] = &cartv1.Item{
			Sku:   item.SKU,
			Count: item.Count,
			Name:  item.Name,
			Price: item.Price,
		}
	}

	return stream.Send(resp)
}



