package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/port"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/usecase/cart/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCartService_ListCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	cartRepository := mocks.NewMockcartRepository(ctrl)
	productClient := mocks.NewMockproductClient(ctrl)
	lomsClient := mocks.NewMocklomsClient(ctrl)

	srv := NewCartService(cartRepository, productClient, lomsClient)

	userID := int64(1)

	cartItems := []entity.CartItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	productsInfo := []*port.ProductInfo{
		{
			Name:  "sneakers",
			Price: 500,
		},
		{
			Name:  "shirt",
			Price: 300,
		},
	}

	cartRepository.EXPECT().
		ListCart(gomock.Any(), userID).
		Return(cartItems, nil)

	productClient.EXPECT().
		GetProductInfoList(gomock.Any(), []uint32{100, 200}).
		Return(productsInfo, nil)

	// Act
	products, totalPrice, err := srv.ListCart(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	require.EqualValues(t, 1300, totalPrice)
	require.Len(t, products, 2)

	require.EqualValues(t, 100, products[0].SKU)
	require.EqualValues(t, 2, products[0].Count)
	require.Equal(t, "sneakers", products[0].Name)
	require.EqualValues(t, 500, products[0].Price)

	require.EqualValues(t, 200, products[1].SKU)
	require.EqualValues(t, 1, products[1].Count)
	require.Equal(t, "shirt", products[1].Name)
	require.EqualValues(t, 300, products[1].Price)
}

func TestCartService_ListCart_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo error")
	productErr := errors.New("product client error")

	tests := []struct {
		name        string
		repoItems   []entity.CartItem
		repoErr     error
		productErr  error
		expectedErr error
	}{
		{
			name:        "repository error",
			repoErr:     repoErr,
			expectedErr: repoErr,
		},
		{
			name:        "empty cart",
			repoItems:   []entity.CartItem{},
			expectedErr: entity.ErrCartIsEmpty,
		},
		{
			name: "product client error",
			repoItems: []entity.CartItem{
				{
					SKU:   100,
					Count: 2,
				},
			},
			productErr:  productErr,
			expectedErr: productErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			cartRepository := mocks.NewMockcartRepository(ctrl)
			productClient := mocks.NewMockproductClient(ctrl)
			lomsClient := mocks.NewMocklomsClient(ctrl)

			srv := NewCartService(cartRepository, productClient, lomsClient)

			userID := int64(1)

			cartRepository.EXPECT().
				ListCart(gomock.Any(), userID).
				Return(tt.repoItems, tt.repoErr)

			if tt.repoErr == nil && len(tt.repoItems) > 0 {
				productClient.EXPECT().
					GetProductInfoList(gomock.Any(), []uint32{100}).
					Return(nil, tt.productErr)
			}

			products, totalPrice, err := srv.ListCart(context.Background(), userID)

			require.Error(t, err)
			require.Nil(t, products)
			require.EqualValues(t, 0, totalPrice)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestCartService_ClearCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	cartRepository := mocks.NewMockcartRepository(ctrl)
	productClient := mocks.NewMockproductClient(ctrl)
	lomsClient := mocks.NewMocklomsClient(ctrl)

	srv := NewCartService(cartRepository, productClient, lomsClient)

	userID := int64(1)

	cartRepository.EXPECT().
		ClearCart(gomock.Any(), userID).
		Return(nil)

	// Act
	err := srv.ClearCart(context.Background(), userID)

	// Assert
	require.NoError(t, err)
}

func TestCartService_ClearCart_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	cartRepository := mocks.NewMockcartRepository(ctrl)
	productClient := mocks.NewMockproductClient(ctrl)
	lomsClient := mocks.NewMocklomsClient(ctrl)

	srv := NewCartService(cartRepository, productClient, lomsClient)

	userID := int64(1)
	expectedErr := errors.New("clear cart error")

	cartRepository.EXPECT().
		ClearCart(gomock.Any(), userID).
		Return(expectedErr)

	// Act
	err := srv.ClearCart(context.Background(), userID)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestCartService_CheckoutCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	cartRepository := mocks.NewMockcartRepository(ctrl)
	productClient := mocks.NewMockproductClient(ctrl)
	lomsClient := mocks.NewMocklomsClient(ctrl)

	srv := NewCartService(cartRepository, productClient, lomsClient)

	userID := int64(1)
	orderID := int64(777)

	cartItems := []entity.CartItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	orderItems := []port.Item{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	cartRepository.EXPECT().
		ListCart(gomock.Any(), userID).
		Return(cartItems, nil)

	lomsClient.EXPECT().
		CreateOrder(gomock.Any(), userID, orderItems).
		Return(orderID, nil)

	cartRepository.EXPECT().
		ClearCart(gomock.Any(), userID).
		Return(nil)

	// Act
	gotOrderID, err := srv.CheckoutCart(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	require.Equal(t, orderID, gotOrderID)
}

func TestCartService_CheckoutCart_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo error")
	lomsErr := errors.New("loms error")
	clearErr := errors.New("clear cart error")

	tests := []struct {
		name        string
		repoItems   []entity.CartItem
		repoErr     error
		lomsErr     error
		clearErr    error
		expectedErr error
	}{
		{
			name:        "repository error",
			repoErr:     repoErr,
			expectedErr: repoErr,
		},
		{
			name:        "empty cart",
			repoItems:   []entity.CartItem{},
			expectedErr: entity.ErrCartIsEmpty,
		},
		{
			name: "loms create order error",
			repoItems: []entity.CartItem{
				{
					SKU:   100,
					Count: 2,
				},
			},
			lomsErr:     lomsErr,
			expectedErr: lomsErr,
		},
		{
			name: "clear cart error",
			repoItems: []entity.CartItem{
				{
					SKU:   100,
					Count: 2,
				},
			},
			clearErr:    clearErr,
			expectedErr: clearErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			cartRepository := mocks.NewMockcartRepository(ctrl)
			productClient := mocks.NewMockproductClient(ctrl)
			lomsClient := mocks.NewMocklomsClient(ctrl)

			srv := NewCartService(cartRepository, productClient, lomsClient)

			userID := int64(1)

			cartRepository.EXPECT().
				ListCart(gomock.Any(), userID).
				Return(tt.repoItems, tt.repoErr)

			if tt.repoErr == nil && len(tt.repoItems) > 0 {
				orderItems := []port.Item{
					{
						SKU:   tt.repoItems[0].SKU,
						Count: tt.repoItems[0].Count,
					},
				}

				lomsClient.EXPECT().
					CreateOrder(gomock.Any(), userID, orderItems).
					Return(int64(0), tt.lomsErr)

				if tt.lomsErr == nil {
					cartRepository.EXPECT().
						ClearCart(gomock.Any(), userID).
						Return(tt.clearErr)
				}
			}

			orderID, err := srv.CheckoutCart(context.Background(), userID)

			require.Error(t, err)
			require.EqualValues(t, 0, orderID)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}



