package controller

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/controller/cart"
	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
)

type API struct {
	cartServer cartv1.CartServer
	logger     *zap.Logger
}

func New(itemService cart.ItemService, cartService cart.CartService, logger *zap.Logger) *API {
	return &API{
		cartServer: cart.NewCartServer(itemService, cartService, logger),
		logger:     logger,
	}
}

func (a *API) Register(grpcServer *grpc.Server) {
	cartv1.RegisterCartServer(grpcServer, a.cartServer)
}
