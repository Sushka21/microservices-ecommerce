package item

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/port"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/usecase/item/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestItemService_AddItem_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	cartRepository := mocks.NewMockcartRepository(ctrl)
	productClient := mocks.NewMockproductClient(ctrl)
	lomsClient := mocks.NewMocklomsClient(ctrl)

	srv := NewItemService(cartRepository, productClient, lomsClient)

	userID := int64(1)
	sku := uint32(100)
	count := uint32(2)

	productClient.EXPECT().
		GetProductInfo(gomock.Any(), sku).
		Return(&port.ProductInfo{
			Name:  "sneakers",
			Price: 500,
		}, nil)

	lomsClient.EXPECT().
		GetStocks(gomock.Any(), sku).
		Return(uint64(10), nil)

	cartRepository.EXPECT().
		AddItem(gomock.Any(), userID, entity.CartItem{
			SKU:   sku,
			Count: count,
		}).
		Return(nil)

	// Act
	err := srv.AddItem(context.Background(), userID, sku, count)

	// Assert
	require.NoError(t, err)
}

func TestItemService_AddItem_Err_Gomock(t *testing.T) {
	t.Parallel()

	productErr := errors.New("product error")
	lomsErr := errors.New("loms error")
	repoErr := errors.New("repo error")

	tests := []struct {
		name        string
		userID      int64
		sku         uint32
		count       uint32
		available   uint64
		productErr  error
		lomsErr     error
		repoErr     error
		expectedErr error
	}{
		{
			name:        "product client error",
			userID:      1,
			sku:         100,
			count:       2,
			productErr:  productErr,
			expectedErr: productErr,
		},
		{
			name:        "loms get stocks error",
			userID:      1,
			sku:         100,
			count:       2,
			lomsErr:     lomsErr,
			expectedErr: lomsErr,
		},
		{
			name:        "insufficient stock",
			userID:      1,
			sku:         100,
			count:       20,
			available:   10,
			expectedErr: entity.ErrInsufficientStock,
		},
		{
			name:        "repository add item error",
			userID:      1,
			sku:         100,
			count:       2,
			available:   10,
			repoErr:     repoErr,
			expectedErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			// Arrange
			cartRepository := mocks.NewMockcartRepository(ctrl)
			productClient := mocks.NewMockproductClient(ctrl)
			lomsClient := mocks.NewMocklomsClient(ctrl)

			srv := NewItemService(cartRepository, productClient, lomsClient)

			productClient.EXPECT().
				GetProductInfo(gomock.Any(), tt.sku).
				Return(&port.ProductInfo{
					Name:  "sneakers",
					Price: 500,
				}, tt.productErr)

			if tt.productErr == nil {
				lomsClient.EXPECT().
					GetStocks(gomock.Any(), tt.sku).
					Return(tt.available, tt.lomsErr)
			}

			if tt.productErr == nil &&
				tt.lomsErr == nil &&
				uint64(tt.count) <= tt.available {
				cartRepository.EXPECT().
					AddItem(gomock.Any(), tt.userID, entity.CartItem{
						SKU:   tt.sku,
						Count: tt.count,
					}).
					Return(tt.repoErr)
			}

			// Act
			err := srv.AddItem(context.Background(), tt.userID, tt.sku, tt.count)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestItemService_DeleteItem_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	cartRepository := mocks.NewMockcartRepository(ctrl)
	productClient := mocks.NewMockproductClient(ctrl)
	lomsClient := mocks.NewMocklomsClient(ctrl)

	srv := NewItemService(cartRepository, productClient, lomsClient)

	userID := int64(1)
	sku := uint32(100)

	cartRepository.EXPECT().
		DeleteItem(gomock.Any(), userID, sku).
		Return(nil)

	// Act
	err := srv.DeleteItem(context.Background(), userID, sku)

	// Assert
	require.NoError(t, err)
}

func TestItemService_DeleteItem_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo error")

	tests := []struct {
		name        string
		userID      int64
		sku         uint32
		repoErr     error
		expectedErr error
	}{
		{
			name:        "repository delete item error",
			userID:      1,
			sku:         100,
			repoErr:     repoErr,
			expectedErr: repoErr,
		},
		{
			name:        "item not found",
			userID:      1,
			sku:         200,
			repoErr:     entity.ErrItemNotFound,
			expectedErr: entity.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			// Arrange
			cartRepository := mocks.NewMockcartRepository(ctrl)
			productClient := mocks.NewMockproductClient(ctrl)
			lomsClient := mocks.NewMocklomsClient(ctrl)

			srv := NewItemService(cartRepository, productClient, lomsClient)

			cartRepository.EXPECT().
				DeleteItem(gomock.Any(), tt.userID, tt.sku).
				Return(tt.repoErr)

			// Act
			err := srv.DeleteItem(context.Background(), tt.userID, tt.sku)

			// Assert
			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
