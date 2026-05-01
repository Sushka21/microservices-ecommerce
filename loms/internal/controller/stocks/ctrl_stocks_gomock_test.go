package stocks

import (
	"context"
	"errors"
	"testing"

	// "github.com/golang/mock/gomock"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/stocks/mocks"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStocksServer_GetStock_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	stocksService := mocks.NewMockStocksService(ctrl)
	srv := NewStocksServer(stocksService, &zap.Logger{})

	req := &stocksv1.GetStockRequest{
		Sku: 100,
	}

	stocksService.EXPECT().
		GetStock(gomock.Any(), req.Sku).
		Return(uint64(17), nil)

	resp, err := srv.GetStock(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(17), resp.Count)
}

func TestStocksServer_GetStock_Err_Gomock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "product not found",
			serviceError: entity.ErrProductNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "unknown error",
			serviceError: errors.New("err"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			stocksService := mocks.NewMockStocksService(ctrl)
			srv := NewStocksServer(stocksService, &zap.Logger{})

			req := &stocksv1.GetStockRequest{
				Sku: 100,
			}

			stocksService.EXPECT().
				GetStock(gomock.Any(), req.Sku).
				Return(uint64(0), tt.serviceError)

			resp, err := srv.GetStock(context.Background(), req)

			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestStocksServer_SetStock_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	stocksService := mocks.NewMockStocksService(ctrl)
	srv := NewStocksServer(stocksService, &zap.Logger{})

	req := &stocksv1.SetStockRequest{
		Sku:   100,
		Count: 25,
	}

	stocksService.EXPECT().
		SetStock(gomock.Any(), req.Sku, req.Count).
		Return(nil)

	resp, err := srv.SetStock(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestStocksServer_SetStock_Err_Gomock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "product not found",
			serviceError: entity.ErrProductNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "unknown error",
			serviceError: errors.New("err"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			stocksService := mocks.NewMockStocksService(ctrl)
			srv := NewStocksServer(stocksService, &zap.Logger{})

			req := &stocksv1.SetStockRequest{
				Sku:   100,
				Count: 25,
			}

			stocksService.EXPECT().
				SetStock(gomock.Any(), req.Sku, req.Count).
				Return(tt.serviceError)

			resp, err := srv.SetStock(context.Background(), req)

			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}



