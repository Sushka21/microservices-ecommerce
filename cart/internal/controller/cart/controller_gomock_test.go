package cart

import (
	"context"
	"errors"
	"testing"

	// "github.com/golang/mock/gomock"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/controller/cart/mocks"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCartServer_AddItem_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	// Arrange
	itemService := mocks.NewMockItemService(ctrl)
	cartService := mocks.NewMockCartService(ctrl)

	srv := NewCartServer(itemService, cartService, zap.NewNop())
	req := &cartv1.AddItemRequest{
		UserId: 1,
		Sku:    100,
		Count:  2,
	}

	itemService.EXPECT().
		AddItem(gomock.Any(), req.UserId, req.Sku, req.Count).
		Return(nil)
	// Act
	resp, err := srv.AddItem(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCartServer_AddItem_err_Gomock(t *testing.T) {
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
			name:         "insufficient stock",
			serviceError: entity.ErrInsufficientStock,
			expectedCode: codes.FailedPrecondition,
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

			itemService := mocks.NewMockItemService(ctrl)
			cartService := mocks.NewMockCartService(ctrl)

			srv := NewCartServer(itemService, cartService, zap.NewNop())

			req := &cartv1.AddItemRequest{
				UserId: 1,
				Sku:    100,
				Count:  2,
			}

			itemService.EXPECT().
				AddItem(gomock.Any(), req.UserId, req.Sku, req.Count).
				Return(tt.serviceError)

			// Act
			resp, err := srv.AddItem(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestCartServer_CheckoutCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	// Arrange
	itemService := mocks.NewMockItemService(ctrl)
	cartService := mocks.NewMockCartService(ctrl)

	srv := NewCartServer(itemService, cartService, zap.NewNop())
	req := &cartv1.CheckoutCartRequest{
		UserId: 1,
	}
	cartService.EXPECT().CheckoutCart(gomock.Any(), req.UserId).Return(int64(3), nil)
	// Act
	resp, err := srv.CheckoutCart(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.EqualValues(t, resp.OrderId, 3)
}

func TestCartServer_Checkout_err_Gomock(t *testing.T) {
	t.Parallel()
	// Arrange
	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{

		{
			name:         "empty cart",
			serviceError: entity.ErrCartIsEmpty,
			expectedCode: codes.FailedPrecondition,
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

			itemService := mocks.NewMockItemService(ctrl)
			cartService := mocks.NewMockCartService(ctrl)

			srv := NewCartServer(itemService, cartService, zap.NewNop())

			req := &cartv1.CheckoutCartRequest{
				UserId: 1,
			}

			cartService.EXPECT().CheckoutCart(gomock.Any(), req.UserId).
				Return(int64(0), tt.serviceError)

			// Act
			resp, err := srv.CheckoutCart(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestCartServer_ClearCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	// Arrange
	itemService := mocks.NewMockItemService(ctrl)
	cartService := mocks.NewMockCartService(ctrl)

	srv := NewCartServer(itemService, cartService, zap.NewNop())
	req := &cartv1.ClearCartRequest{
		UserId: 1,
	}

	cartService.EXPECT().
		ClearCart(gomock.Any(), req.UserId).
		Return(nil)

	// Act
	resp, err := srv.ClearCart(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCartServer_ClearCart_err_Gomock(t *testing.T) {
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

			itemService := mocks.NewMockItemService(ctrl)
			cartService := mocks.NewMockCartService(ctrl)

			srv := NewCartServer(itemService, cartService, zap.NewNop())

			req := &cartv1.ClearCartRequest{
				UserId: 1,
			}

			cartService.EXPECT().
				ClearCart(gomock.Any(), req.UserId).
				Return(tt.serviceError)

			// Act
			resp, err := srv.ClearCart(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestCartServer_DeleteItem_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	// Arrange
	itemService := mocks.NewMockItemService(ctrl)
	cartService := mocks.NewMockCartService(ctrl)

	srv := NewCartServer(itemService, cartService, zap.NewNop())
	req := &cartv1.DeleteItemRequest{
		UserId: 1,
		Sku:    100,
	}

	itemService.EXPECT().
		DeleteItem(gomock.Any(), req.UserId, req.Sku).
		Return(nil)

	// Act
	resp, err := srv.DeleteItem(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCartServer_DeleteItem_err_Gomock(t *testing.T) {
	t.Parallel()

	// Arrange
	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "item not found",
			serviceError: entity.ErrItemNotFound,
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

			itemService := mocks.NewMockItemService(ctrl)
			cartService := mocks.NewMockCartService(ctrl)

			srv := NewCartServer(itemService, cartService, zap.NewNop())

			req := &cartv1.DeleteItemRequest{
				UserId: 1,
				Sku:    100,
			}

			itemService.EXPECT().
				DeleteItem(gomock.Any(), req.UserId, req.Sku).
				Return(tt.serviceError)

			// Act
			resp, err := srv.DeleteItem(context.Background(), req)

			// Assert
			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestCartServer_ListCart_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	// Arrange
	itemService := mocks.NewMockItemService(ctrl)
	cartService := mocks.NewMockCartService(ctrl)

	srv := NewCartServer(itemService, cartService, zap.NewNop())
	req := &cartv1.ListCartRequest{
		UserId: 1,
	}

	stream := &mockListCartServer{
		ctx: context.Background(),
	}

	items := []entity.Product{
		{
			SKU:   100,
			Count: 2,
			Name:  "sneakers",
			Price: 500,
		},
		{
			SKU:   200,
			Count: 1,
			Name:  "shirt",
			Price: 300,
		},
	}

	cartService.EXPECT().
		ListCart(gomock.Any(), req.UserId).
		Return(items, uint32(800), nil)

	// Act
	err := srv.ListCart(req, stream)

	// Assert
	require.NoError(t, err)
	require.Len(t, stream.cart, 1)

	resp := stream.cart[0]
	require.NotNil(t, resp)
	require.EqualValues(t, 800, resp.TotalPrice)
	require.Len(t, resp.Items, 2)

	require.EqualValues(t, 100, resp.Items[0].Sku)
	require.EqualValues(t, 2, resp.Items[0].Count)
	require.Equal(t, "sneakers", resp.Items[0].Name)
	require.EqualValues(t, 500, resp.Items[0].Price)

	require.EqualValues(t, 200, resp.Items[1].Sku)
	require.EqualValues(t, 1, resp.Items[1].Count)
	require.Equal(t, "shirt", resp.Items[1].Name)
	require.EqualValues(t, 300, resp.Items[1].Price)
}

func TestCartServer_ListCart_err_Gomock(t *testing.T) {
	t.Parallel()

	// Arrange
	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "empty cart",
			serviceError: entity.ErrCartIsEmpty,
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

			itemService := mocks.NewMockItemService(ctrl)
			cartService := mocks.NewMockCartService(ctrl)

			srv := NewCartServer(itemService, cartService, zap.NewNop())

			req := &cartv1.ListCartRequest{
				UserId: 1,
			}

			stream := &mockListCartServer{
				ctx: context.Background(),
			}

			cartService.EXPECT().
				ListCart(gomock.Any(), req.UserId).
				Return([]entity.Product(nil), uint32(0), tt.serviceError)

			// Act
			err := srv.ListCart(req, stream)

			if errors.Is(tt.serviceError, entity.ErrCartIsEmpty) {
				require.NoError(t, err)
				require.Len(t, stream.cart, 0)
				return
			}

			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

type mockListCartServer struct {
	cartv1.Cart_ListCartServer
	ctx     context.Context
	cart    []*cartv1.ListCartResponse
	sendErr error
}

func (m *mockListCartServer) Context() context.Context {
	return m.ctx
}

func (m *mockListCartServer) Send(resp *cartv1.ListCartResponse) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.cart = append(m.cart, resp)
	return nil
}



