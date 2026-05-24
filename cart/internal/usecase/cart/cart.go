package cart

import (
	"context"
	"fmt"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/port"
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mocks.go -package=mocks
type (
	cartRepository interface {
		ListCart(ctx context.Context, userID int64) ([]entity.CartItem, error)
		ClearCart(ctx context.Context, userID int64) error
	}

	productClient interface {
		GetProductInfoList(ctx context.Context, skus []uint32) ([]*port.ProductInfo, error)
	}

	lomsClient interface {
		CreateOrder(ctx context.Context, userID int64, items []port.Item) (int64, error)
	}
)

type cartService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewCartService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient) *cartService {
	return &cartService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
	}
}

func (s *cartService) ListCart(ctx context.Context, userID int64) ([]entity.Product, uint32, error) {
	items, err := s.cartRepository.ListCart(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("list cart items for user_id=%d: %w", userID, err)
	}

	skus := make([]uint32, len(items))
	for i := range items {
		skus[i] = items[i].SKU
	}

	if len(skus) == 0 {
		return nil, 0, entity.ErrCartIsEmpty
	}

	productsInfo, err := s.productClient.GetProductInfoList(ctx, skus)
	if err != nil {
		return nil, 0, fmt.Errorf("get product info list failed for user_id=%d: %w", userID, err)
	}

	product := make([]entity.Product, len(items))
	var totalPrice uint32
	for i := range product {
		product[i] = entity.Product{
			Name:  productsInfo[i].Name,
			Price: productsInfo[i].Price,
			SKU:   skus[i],
			Count: items[i].Count,
		}
		totalPrice += product[i].Price * (product[i].Count)
	}
	return product, totalPrice, nil
}

func (s *cartService) ClearCart(ctx context.Context, userID int64) error {
	if err := s.cartRepository.ClearCart(ctx, userID); err != nil {
		return fmt.Errorf("clear cart for user_id=%d: %w", userID, err)
	}
	return nil
}

func (s *cartService) CheckoutCart(ctx context.Context, userID int64) (int64, error) {
	items, err := s.cartRepository.ListCart(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("list cart items for user_id=%d: %w", userID, err)
	}

	if len(items) == 0 {
		return 0, entity.ErrCartIsEmpty
	}

	orderItems := make([]port.Item, len(items))
	for i := range items {
		orderItems[i] = port.Item{
			SKU:   items[i].SKU,
			Count: items[i].Count,
		}
	}

	orderID, err := s.lomsClient.CreateOrder(ctx, userID, orderItems)
	if err != nil {
		return 0, fmt.Errorf("loms CreateOrder failed for user_id=%d, items=%v: %w", userID, orderItems, err)
	}

	if err := s.cartRepository.ClearCart(ctx, userID); err != nil {
		return 0, fmt.Errorf("clear cart: %w", err)
	}

	return orderID, nil
}
