package product

import (
	"context"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	"go.uber.org/zap"
)

//go:generate mockgen -source=product.go -destination=mocks/product_mocks.go -package=mocks
type (
	//nolint:revive // ProductService name is intentionally explicit because this package has multiple service interfaces.
	ProductService interface {
		CreateProduct(ctx context.Context, name string, price uint32) (uint32, error)
		GetProduct(ctx context.Context, sku uint32) (entity.ProductInfo, error)
		ListProduct(ctx context.Context, skus []uint32) ([]entity.ProductInfo, error)
	}
)

type productServer struct {
	productv1.UnimplementedProductServiceServer
	productService ProductService
	logger         *zap.Logger
}

func NewProductServer(productService ProductService, logger *zap.Logger) *productServer {
	return &productServer{
		productService: productService,
		logger:         logger,
	}
}



