package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/repository/cart/mocks"
	sqlcCart "github.com/Sushka21/microservices-ecommerce/cart/internal/repository/cart/sqlc"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPostgresRepository_AddItem_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)
	item := entity.CartItem{
		SKU:   100,
		Count: 2,
	}

	querier.EXPECT().
		InsertItem(gomock.Any(), sqlcCart.InsertItemParams{
			UserID: userID,
			Sku:    int64(item.SKU),
			Count:  int64(item.Count),
		}).
		Return(nil)

	// Act
	err := repo.AddItem(context.Background(), userID, item)

	// Assert
	require.NoError(t, err)
}

func TestPostgresRepository_AddItem_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)
	item := entity.CartItem{
		SKU:   100,
		Count: 2,
	}
	expectedErr := errors.New("insert item error")

	querier.EXPECT().
		InsertItem(gomock.Any(), sqlcCart.InsertItemParams{
			UserID: userID,
			Sku:    int64(item.SKU),
			Count:  int64(item.Count),
		}).
		Return(expectedErr)

	// Act
	err := repo.AddItem(context.Background(), userID, item)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestPostgresRepository_DeleteItem_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)
	sku := uint32(100)

	querier.EXPECT().
		DeleteItemBySku(gomock.Any(), sqlcCart.DeleteItemBySkuParams{
			UserID: userID,
			Sku:    int64(sku),
		}).
		Return(nil)

	// Act
	err := repo.DeleteItem(context.Background(), userID, sku)

	// Assert
	require.NoError(t, err)
}

func TestPostgresRepository_DeleteItem_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)
	sku := uint32(100)
	expectedErr := errors.New("delete item error")

	querier.EXPECT().
		DeleteItemBySku(gomock.Any(), sqlcCart.DeleteItemBySkuParams{
			UserID: userID,
			Sku:    int64(sku),
		}).
		Return(expectedErr)

	// Act
	err := repo.DeleteItem(context.Background(), userID, sku)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestPostgresRepository_ListCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)

	querier.EXPECT().
		ListCartByUserId(gomock.Any(), userID).
		Return([]sqlcCart.ListCartByUserIdRow{
			{
				Sku:   100,
				Count: 2,
			},
			{
				Sku:   200,
				Count: 1,
			},
		}, nil)

	// Act
	items, err := repo.ListCart(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	require.Len(t, items, 2)

	require.EqualValues(t, 100, items[0].SKU)
	require.EqualValues(t, 2, items[0].Count)

	require.EqualValues(t, 200, items[1].SKU)
	require.EqualValues(t, 1, items[1].Count)
}

func TestPostgresRepository_ListCart_Empty_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)

	querier.EXPECT().
		ListCartByUserId(gomock.Any(), userID).
		Return([]sqlcCart.ListCartByUserIdRow{}, nil)

	// Act
	items, err := repo.ListCart(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestPostgresRepository_ListCart_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)
	expectedErr := errors.New("list cart error")

	querier.EXPECT().
		ListCartByUserId(gomock.Any(), userID).
		Return(nil, expectedErr)

	// Act
	items, err := repo.ListCart(context.Background(), userID)

	// Assert
	require.Error(t, err)
	require.Nil(t, items)
	require.ErrorIs(t, err, expectedErr)
}

func TestPostgresRepository_ClearCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)

	querier.EXPECT().
		ClearCart(gomock.Any(), userID).
		Return(nil)

	// Act
	err := repo.ClearCart(context.Background(), userID)

	// Assert
	require.NoError(t, err)
}

func TestPostgresRepository_ClearCart_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	userID := int64(1)
	expectedErr := errors.New("clear cart error")

	querier.EXPECT().
		ClearCart(gomock.Any(), userID).
		Return(expectedErr)

	// Act
	err := repo.ClearCart(context.Background(), userID)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}



