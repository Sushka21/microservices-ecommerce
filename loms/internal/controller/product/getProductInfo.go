package product

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServer) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	productInfo, err := s.productService.GetProduct(ctx, req.GetSku())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		default:
			s.logger.Error(
				"failed to get product to product",
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Internal, "internal error")
		}
	}
	return &productv1.GetProductResponse{
		Name:  productInfo.Name,
		Price: productInfo.Price,
	}, nil
}
