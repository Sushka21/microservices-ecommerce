package grpc

import (
	"context"
	"fmt"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/port"
	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type productClient struct {
	client productv1.ProductServiceClient
}

func NewProductClient(client productv1.ProductServiceClient) *productClient {
	return &productClient{
		client: client,
	}
}

func (c *productClient) GetProductInfo(ctx context.Context, sku uint32) (*port.ProductInfo, error) {
	resp, err := c.client.GetProduct(ctx, &productv1.GetProductRequest{
		Sku: sku,
	})

	if err != nil {
		return nil, mapProductError(err)
	}

	return &port.ProductInfo{
		Name:  resp.GetName(),
		Price: resp.GetPrice(),
	}, nil
}

func (c *productClient) GetProductInfoList(ctx context.Context, skus []uint32) ([]*port.ProductInfo, error) {
	resp, err := c.client.ListProduct(ctx, &productv1.ListProductsRequest{
		Skus: skus,
	})

	if err != nil {
		return nil, mapProductError(err)
	}

	productsInfo := make([]*port.ProductInfo, len(resp.GetProducts()))
	for i, p := range resp.GetProducts() {
		productsInfo[i] = &port.ProductInfo{
			Name:  p.GetName(),
			Price: p.GetPrice(),
		}
	}
	return productsInfo, nil
}

func mapProductError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("product grpc call failed: %w", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %v", entity.ErrProductNotFound, err)
	default:
		return fmt.Errorf("product grpc call failed: %w", err)
	}
}



