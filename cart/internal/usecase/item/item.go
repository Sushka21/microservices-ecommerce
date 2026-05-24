package item

import (
	"context"
	"fmt"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/port"
)

//go:generate mockgen -source=item.go -destination=mocks/item_mocks.go -package=mocks
type (
	cartRepository interface {
		AddItem(ctx context.Context, userID int64, item entity.CartItem) error
		DeleteItem(ctx context.Context, userID int64, sku uint32) error
	}

	productClient interface {
		GetProductInfo(ctx context.Context, sku uint32) (*port.ProductInfo, error)
	}

	lomsClient interface {
		GetStocks(ctx context.Context, sku uint32) (uint64, error)
	}
)

type itemService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewItemService(cartRepository cartRepository, produproductClient productClient, lomslomsClient lomsClient) *itemService {
	return &itemService{
		cartRepository: cartRepository,
		productClient:  produproductClient,
		lomsClient:     lomslomsClient,
	}
}

func (s *itemService) AddItem(ctx context.Context, userID int64, sku, count uint32) error {
	_, err := s.productClient.GetProductInfo(ctx, sku)
	if err != nil {
		return err
	}
	available, err := s.lomsClient.GetStocks(ctx, sku)
	if err != nil {
		return fmt.Errorf("loms GetStocks failed for user_id=%d sku=%d: %w", userID, sku, err)
	}
	if uint64(count) > available {
		return fmt.Errorf("insufficient stock, requested %d, got %d: %w",
			count,
			available,
			entity.ErrInsufficientStock,
		)
	}
	return s.cartRepository.AddItem(ctx, userID, entity.CartItem{SKU: sku, Count: count})
}

func (s *itemService) DeleteItem(ctx context.Context, userID int64, sku uint32) error {
	if err := s.cartRepository.DeleteItem(ctx, userID, sku); err != nil {
		return fmt.Errorf("delete item from cart for user_id=%d sku=%d: %w", userID, sku, err)
	}
	return nil
}
