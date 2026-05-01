package product

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/usercase/product/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProductService_CreateProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	productRepository := mocks.NewMockproductRepository(ctrl)

	srv := NewProductService(productRepository)

	ctx := context.Background()
	name := "sneakers"
	price := uint32(500)
	expectedSKU := uint32(100)

	productRepository.EXPECT().
		CreateProduct(gomock.Any(), entity.ProductInfo{
			Name:  name,
			Price: price,
			Count: 1,
		}).
		Return(expectedSKU, nil)

	// Act
	sku, err := srv.CreateProduct(ctx, name, price)

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedSKU, sku)
}

func TestProductService_CreateProduct_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	productRepository := mocks.NewMockproductRepository(ctrl)

	srv := NewProductService(productRepository)

	ctx := context.Background()
	name := "sneakers"
	price := uint32(500)
	expectedErr := errors.New("create product error")

	productRepository.EXPECT().
		CreateProduct(gomock.Any(), entity.ProductInfo{
			Name:  name,
			Price: price,
			Count: 1,
		}).
		Return(uint32(0), expectedErr)

	// Act
	sku, err := srv.CreateProduct(ctx, name, price)

	// Assert
	require.Error(t, err)
	require.EqualValues(t, 0, sku)
	require.ErrorIs(t, err, expectedErr)
}

func TestProductService_GetProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	productRepository := mocks.NewMockproductRepository(ctrl)

	srv := NewProductService(productRepository)

	ctx := context.Background()
	sku := uint32(100)

	expectedProduct := entity.ProductInfo{
		Name:  "sneakers",
		Price: 500,
		Count: 10,
	}

	productRepository.EXPECT().
		GetProductBySKU(gomock.Any(), sku).
		Return(expectedProduct, nil)

	// Act
	productInfo, err := srv.GetProduct(ctx, sku)

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedProduct, productInfo)
}

func TestProductService_GetProduct_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	productRepository := mocks.NewMockproductRepository(ctrl)

	srv := NewProductService(productRepository)

	ctx := context.Background()
	sku := uint32(100)
	expectedErr := errors.New("get product error")

	productRepository.EXPECT().
		GetProductBySKU(gomock.Any(), sku).
		Return(entity.ProductInfo{}, expectedErr)

	// Act
	productInfo, err := srv.GetProduct(ctx, sku)

	// Assert
	require.Error(t, err)
	require.Equal(t, entity.ProductInfo{}, productInfo)
	require.ErrorIs(t, err, expectedErr)
}

func TestProductService_ListProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	productRepository := mocks.NewMockproductRepository(ctrl)

	srv := NewProductService(productRepository)

	ctx := context.Background()

	skus := []uint32{100, 200}

	firstProduct := entity.ProductInfo{
		Name:  "sneakers",
		Price: 500,
		Count: 10,
	}

	secondProduct := entity.ProductInfo{
		Name:  "shirt",
		Price: 300,
		Count: 5,
	}

	gomock.InOrder(
		productRepository.EXPECT().
			GetProductBySKU(gomock.Any(), uint32(100)).
			Return(firstProduct, nil),

		productRepository.EXPECT().
			GetProductBySKU(gomock.Any(), uint32(200)).
			Return(secondProduct, nil),
	)

	// Act
	products, err := srv.ListProduct(ctx, skus)

	// Assert
	require.NoError(t, err)
	require.Len(t, products, 2)
	require.Equal(t, firstProduct, products[0])
	require.Equal(t, secondProduct, products[1])
}

func TestProductService_ListProduct_Empty_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	productRepository := mocks.NewMockproductRepository(ctrl)

	srv := NewProductService(productRepository)

	ctx := context.Background()
	skus := []uint32{}

	// Act
	products, err := srv.ListProduct(ctx, skus)

	// Assert
	require.NoError(t, err)
	require.Empty(t, products)
}

func TestProductService_ListProduct_Err_Gomock(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("get product error")

	tests := []struct {
		name        string
		skus        []uint32
		setupMock   func(productRepository *mocks.MockproductRepository)
		expectedErr error
	}{
		{
			name: "first product error",
			skus: []uint32{100, 200},
			setupMock: func(productRepository *mocks.MockproductRepository) {
				productRepository.EXPECT().
					GetProductBySKU(gomock.Any(), uint32(100)).
					Return(entity.ProductInfo{}, expectedErr)
			},
			expectedErr: expectedErr,
		},
		{
			name: "second product error",
			skus: []uint32{100, 200},
			setupMock: func(productRepository *mocks.MockproductRepository) {
				productRepository.EXPECT().
					GetProductBySKU(gomock.Any(), uint32(100)).
					Return(entity.ProductInfo{
						Name:  "sneakers",
						Price: 500,
						Count: 10,
					}, nil)

				productRepository.EXPECT().
					GetProductBySKU(gomock.Any(), uint32(200)).
					Return(entity.ProductInfo{}, expectedErr)
			},
			expectedErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			// Arrange
			productRepository := mocks.NewMockproductRepository(ctrl)

			srv := NewProductService(productRepository)

			tt.setupMock(productRepository)

			// Act
			products, err := srv.ListProduct(context.Background(), tt.skus)

			// Assert
			require.Error(t, err)
			require.Nil(t, products)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}



