package controller

import (
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/loms"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/product"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/stocks"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type API struct {
	lomsServer    lomsv1.LomsServer
	productServer productv1.ProductServiceServer
	stocksServer  stocksv1.StocksServer
}

func New(lomsService loms.LomsService, productService product.ProductService, stoksService stocks.StocksService,
	logger *zap.Logger) *API {
	return &API{
		lomsServer:    loms.NewLomsServer(lomsService, logger),
		productServer: product.NewProductServer(productService, logger),
		stocksServer:  stocks.NewStocksServer(stoksService, logger),
	}
}

func (a *API) Register(grpcServer *grpc.Server) {
	lomsv1.RegisterLomsServer(grpcServer, a.lomsServer)
	productv1.RegisterProductServiceServer(grpcServer, a.productServer)
	stocksv1.RegisterStocksServer(grpcServer, a.stocksServer)
}



