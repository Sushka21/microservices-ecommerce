package stocks

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/usercase/stocks/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStocksService_SetStock_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	stocksRepository := mocks.NewMockstocksRepository(ctrl)

	srv := NewStocksService(stocksRepository)

	ctx := context.Background()
	sku := uint32(100)
	count := uint64(10)

	stocksRepository.EXPECT().
		SetCountBySKU(gomock.Any(), sku, count).
		Return(nil)

	// Act
	err := srv.SetStock(ctx, sku, count)

	// Assert
	require.NoError(t, err)
}

func TestStocksService_SetStock_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	stocksRepository := mocks.NewMockstocksRepository(ctrl)

	srv := NewStocksService(stocksRepository)

	ctx := context.Background()
	sku := uint32(100)
	count := uint64(10)
	expectedErr := errors.New("set stock error")

	stocksRepository.EXPECT().
		SetCountBySKU(gomock.Any(), sku, count).
		Return(expectedErr)

	// Act
	err := srv.SetStock(ctx, sku, count)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestStocksService_GetStock_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	stocksRepository := mocks.NewMockstocksRepository(ctrl)

	srv := NewStocksService(stocksRepository)

	ctx := context.Background()
	sku := uint32(100)
	expectedCount := uint64(10)

	stocksRepository.EXPECT().
		GetCountBySKU(gomock.Any(), sku).
		Return(expectedCount, nil)

	// Act
	count, err := srv.GetStock(ctx, sku)

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedCount, count)
}

func TestStocksService_GetStock_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	stocksRepository := mocks.NewMockstocksRepository(ctrl)

	srv := NewStocksService(stocksRepository)

	ctx := context.Background()
	sku := uint32(100)
	expectedErr := errors.New("get stock error")

	stocksRepository.EXPECT().
		GetCountBySKU(gomock.Any(), sku).
		Return(uint64(0), expectedErr)

	// Act
	count, err := srv.GetStock(ctx, sku)

	// Assert
	require.Error(t, err)
	require.EqualValues(t, 0, count)
	require.ErrorIs(t, err, expectedErr)
}



