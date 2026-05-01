package stocks

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s stocksServer) SetStock(ctx context.Context, req *stocksv1.SetStockRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.stocksService.SetStock(ctx, req.GetSku(), req.GetCount()); err != nil {
		switch {
		case errors.Is(err, entity.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		default:
			s.logger.Error(
				"failed to set stock to stock",
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Internal, "internal error")
		}
	}
	return &emptypb.Empty{}, nil
}



