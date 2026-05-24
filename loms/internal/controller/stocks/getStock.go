package stocks

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s stocksServer) GetStock(ctx context.Context, req *stocksv1.GetStockRequest) (*stocksv1.GetStockResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	count, err := s.stocksService.GetStock(ctx, req.GetSku())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		default:
			s.logger.Error(
				"failed to get stock to stock",
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Internal, "internal error")
		}
	}
	return &stocksv1.GetStockResponse{
		Count: count,
	}, nil
}
