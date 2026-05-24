package stocks

import (
	"context"
	"fmt"
)

//go:generate mockgen -source=stocks.go -destination=mocks/stock_mocks.go -package=mocks
type (
	stocksRepository interface {
		GetCountBySKU(ctx context.Context, sku uint32) (uint64, error)
		SetCountBySKU(ctx context.Context, sku uint32, count uint64) error
	}
)

type stocksService struct {
	stocksRepository stocksRepository
}

func NewStocksService(stocksRepository stocksRepository) *stocksService {
	return &stocksService{
		stocksRepository: stocksRepository,
	}
}

func (s *stocksService) SetStock(ctx context.Context, sku uint32, count uint64) error {
	if err := s.stocksRepository.SetCountBySKU(ctx, sku, count); err != nil {
		return fmt.Errorf("set stock for sku=%d, count=%d: %w", sku, count, err)
	}
	return nil
}
func (s *stocksService) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	count, err := s.stocksRepository.GetCountBySKU(ctx, sku)
	if err != nil {
		return 0, fmt.Errorf("get product by sku=%d: %w", sku, err)
	}
	return count, err
}
