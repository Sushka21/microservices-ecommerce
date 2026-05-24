package product

import (
	"context"

	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServer) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	sku, err := s.productService.CreateProduct(ctx, req.GetName(), req.GetPrice())
	if err != nil {
		s.logger.Error(
			"failed to creat product to product",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &productv1.CreateProductResponse{
		Sku: sku,
	}, nil
}
