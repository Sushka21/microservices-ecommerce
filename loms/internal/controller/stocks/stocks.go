package stocks

import (
	"context"

	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"
	"go.uber.org/zap"
)

//go:generate mockgen -source=stocks.go -destination=mocks/stocks_mocks.go -package=mocks
type (
	//nolint:revive // StocksService name is intentionally explicit because this package has multiple service interfaces.
	StocksService interface {
		SetStock(ctx context.Context, sku uint32, count uint64) error
		GetStock(ctx context.Context, sku uint32) (uint64, error)
	}
)

type stocksServer struct {
	stocksv1.UnimplementedStocksServer
	stocksService StocksService
	logger        *zap.Logger
}

func NewStocksServer(stocksService StocksService, logger *zap.Logger) *stocksServer {
	return &stocksServer{
		stocksService: stocksService,
		logger:        logger,
	}
}



