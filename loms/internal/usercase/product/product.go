package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

//go:generate mockgen -source=product.go -destination=mocks/product_mocks.go -package=mocks
type (
	productRepository interface {
		CreateProduct(ctx context.Context, product entity.ProductInfo) (uint32, error)
		GetProductBySKU(ctx context.Context, sku uint32) (entity.ProductInfo, error)
	}
)

type productService struct {
	productRepository productRepository
}

func NewProductService(productRepository productRepository) *productService {
	return &productService{
		productRepository: productRepository,
	}
}

func (s *productService) CreateProduct(ctx context.Context, name string, price uint32) (uint32, error) {
	sku, err := s.productRepository.CreateProduct(ctx, entity.ProductInfo{
		Name:  name,
		Price: price,
		Count: 1,
	})

	if err != nil {
		return 0, fmt.Errorf("create product name=%q price=%d: %w", name, price, err)
	}

	return sku, nil
}

func (s *productService) GetProduct(ctx context.Context, sku uint32) (entity.ProductInfo, error) {
	productInfo, err := s.productRepository.GetProductBySKU(ctx, sku)
	if err != nil {
		return entity.ProductInfo{}, fmt.Errorf("get product by sku=%d: %w", sku, err)
	}

	return productInfo, nil
}

func (s *productService) ListProduct(ctx context.Context, skus []uint32) ([]entity.ProductInfo, error) {
	productInfo := make([]entity.ProductInfo, len(skus))

	for i := range skus {
		p, err := s.GetProduct(ctx, skus[i])
		if err != nil {
			return nil, fmt.Errorf("list products sku=%d index=%d: %w", skus[i], i, err)
		}

		productInfo[i] = p
	}

	return productInfo, nil
}



