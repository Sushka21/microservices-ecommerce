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

func (s *productServer) ListProduct(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	lisProductInfo, err := s.productService.ListProduct(ctx, req.GetSkus())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product  not found")
		default:
			s.logger.Error(
				"failed to list product to product",
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Internal, "internal error")
		}
	}
	products := make([]*productv1.ProductInfo, len(lisProductInfo))
	for i, p := range lisProductInfo {
		products[i] = &productv1.ProductInfo{
			Name:  p.Name,
			Sku:   p.Sku,
			Price: p.Price,
		}
	}
	return &productv1.ListProductsResponse{
		Products: products,
	}, nil
}
