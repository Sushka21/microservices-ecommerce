package grpc

import (
	"context"
	"fmt"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/port"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lomsClient struct {
	lomsClient  lomsv1.LomsClient
	stockslient stocksv1.StocksClient
}

func NewLOMSClient(client lomsv1.LomsClient, stockslient stocksv1.StocksClient) *lomsClient {
	return &lomsClient{
		lomsClient:  client,
		stockslient: stockslient,
	}
}

func (c *lomsClient) GetStocks(ctx context.Context, sku uint32) (uint64, error) {
	resp, err := c.stockslient.GetStock(ctx, &stocksv1.GetStockRequest{
		Sku: sku,
	})

	if err != nil {
		return 0, mapGetStocksError(err)
	}

	return resp.GetCount(), nil
}

func (c *lomsClient) CreateOrder(ctx context.Context, userID int64, items []port.Item) (int64, error) {
	orderItems := make([]*lomsv1.Item, len(items))
	for i, item := range items {
		orderItems[i] = &lomsv1.Item{
			Sku:   item.SKU,
			Count: item.Count,
		}
	}

	resp, err := c.lomsClient.CreateOrder(ctx, &lomsv1.CreateOrderRequest{
		UserId: userID,
		Items:  orderItems,
	})

	if err != nil {
		return 0, mapCreateOrderError(err)
	}

	return resp.GetOrderId(), nil
}

func mapGetStocksError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("get stock grpc failed: %w", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %v", entity.ErrProductNotFound, err)
	default:
		return fmt.Errorf("get stock grpc failed: %w", err)
	}
}

func mapCreateOrderError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("%w: %v", port.ErrOrderCreateFailed, err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %v", entity.ErrProductNotFound, err)
	case codes.FailedPrecondition:
		return fmt.Errorf("%w: %v", entity.ErrInsufficientStock, err)
	default:
		return fmt.Errorf("%w: %v", port.ErrOrderCreateFailed, err)
	}
}



