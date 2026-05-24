package cart

import (
	"context"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	"go.uber.org/zap"
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mocks.go -package=mocks
type ItemService interface {
	AddItem(ctx context.Context, userID int64, sku, count uint32) error
	DeleteItem(ctx context.Context, userID int64, sku uint32) error
}

//nolint:revive // CartService name is intentionally explicit because this package has multiple service interfaces.
type CartService interface {
	ListCart(ctx context.Context, userID int64) (items []entity.Product, totalPrice uint32, err error)
	ClearCart(ctx context.Context, userID int64) error
	CheckoutCart(ctx context.Context, userID int64) (orderID int64, err error)
}

type cartServer struct {
	cartv1.UnimplementedCartServer
	itemService ItemService
	cartService CartService
	logger      *zap.Logger
}

func NewCartServer(
	itemService ItemService,
	cartService CartService,
	logger *zap.Logger,
) *cartServer {
	return &cartServer{
		itemService: itemService,
		cartService: cartService,
		logger:      logger,
	}
}
