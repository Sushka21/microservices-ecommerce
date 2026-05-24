package product

import (
	"context"
	"errors"
	"testing"

	// "github.com/golang/mock/gomock"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/product/mocks"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductServer_CreateProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	// Arrange
	productService := mocks.NewMockProductService(ctrl)

	srv := NewProductServer(productService, zap.NewNop())
	req := &productv1.CreateProductRequest{
		Name:  "apple",
		Price: 500,
	}

	productService.EXPECT().
		CreateProduct(gomock.Any(), req.Name, req.Price).
		Return(uint32(100), nil)

	// Act
	resp, err := srv.CreateProduct(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.EqualValues(t, 100, resp.Sku)
}

func TestProductServer_CreateProduct_err_Gomock(t *testing.T) {
	t.Parallel()

	// Arrange
	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "unknown error",
			serviceError: errors.New("err"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			productService := mocks.NewMockProductService(ctrl)

			srv := NewProductServer(productService, zap.NewNop())

			req := &productv1.CreateProductRequest{
				Name:  "shirt",
				Price: 500,
			}

			productService.EXPECT().
				CreateProduct(gomock.Any(), req.Name, req.Price).
				Return(uint32(0), tt.serviceError)

			// Act
			resp, err := srv.CreateProduct(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestProductServer_GetProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	// Arrange
	productService := mocks.NewMockProductService(ctrl)

	srv := NewProductServer(productService, zap.NewNop())
	req := &productv1.GetProductRequest{
		Sku: 100,
	}

	productService.EXPECT().
		GetProduct(gomock.Any(), req.Sku).
		Return(entity.ProductInfo{
			Sku:   100,
			Name:  "shirt",
			Price: 500,
		}, nil)

	// Act
	resp, err := srv.GetProduct(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "shirt", resp.Name)
	require.EqualValues(t, 500, resp.Price)
}

func TestProductServer_GetProduct_err_Gomock(t *testing.T) {
	t.Parallel()

	// Arrange
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			productService := mocks.NewMockProductService(ctrl)

			srv := NewProductServer(productService, zap.NewNop())

			req := &productv1.GetProductRequest{
				Sku: 100,
			}

			productService.EXPECT().
				GetProduct(gomock.Any(), req.Sku).
				Return(entity.ProductInfo{}, tt.serviceError)

			// Act
			resp, err := srv.GetProduct(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestProductServer_ListProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	// Arrange
	productService := mocks.NewMockProductService(ctrl)

	srv := NewProductServer(productService, zap.NewNop())
	req := &productv1.ListProductsRequest{
		Skus: []uint32{100, 200},
	}

	productService.EXPECT().
		ListProduct(gomock.Any(), req.Skus).
		Return([]entity.ProductInfo{
			{
				Sku:   100,
				Name:  "shirt",
				Price: 500,
			},
			{
				Sku:   200,
				Name:  "sneakers",
				Price: 300,
			},
		}, nil)

	// Act
	resp, err := srv.ListProduct(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Products, 2)

	require.EqualValues(t, 100, resp.Products[0].Sku)
	require.Equal(t, "shirt", resp.Products[0].Name)
	require.EqualValues(t, 500, resp.Products[0].Price)

	require.EqualValues(t, 200, resp.Products[1].Sku)
	require.Equal(t, "sneakers", resp.Products[1].Name)
	require.EqualValues(t, 300, resp.Products[1].Price)
}

func TestProductServer_ListProduct_err_Gomock(t *testing.T) {
	t.Parallel()

	// Arrange
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			productService := mocks.NewMockProductService(ctrl)

			srv := NewProductServer(productService, zap.NewNop())

			req := &productv1.ListProductsRequest{
				Skus: []uint32{100, 200},
			}

			productService.EXPECT().
				ListProduct(gomock.Any(), req.Skus).
				Return(nil, tt.serviceError)

			// Act
			resp, err := srv.ListProduct(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}
