package loms

import (
	"context"
	"errors"
	"testing"
	"time"

	// "github.com/golang/mock/gomock"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller/loms/mocks"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLomsServer_CancelOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	lomsService := mocks.NewMockLomsService(ctrl)
	srv := NewLomsServer(lomsService, zap.NewNop())

	req := &lomsv1.CancelOrderRequest{
		OrderId: 1,
	}

	lomsService.EXPECT().
		CancelOrder(gomock.Any(), req.OrderId).
		Return(nil)

	resp, err := srv.CancelOrder(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestLomsServer_CancelOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "order already cancelled",
			serviceError: entity.ErrOrderCancelled,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "order already paid",
			serviceError: entity.ErrOrderAlreadyPaid,
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

			lomsService := mocks.NewMockLomsService(ctrl)
			srv := NewLomsServer(lomsService, zap.NewNop())

			req := &lomsv1.CancelOrderRequest{
				OrderId: 1,
			}

			lomsService.EXPECT().
				CancelOrder(gomock.Any(), req.OrderId).
				Return(tt.serviceError)

			resp, err := srv.CancelOrder(context.Background(), req)

			require.NotNil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestLomsServer_CreateOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	lomsService := mocks.NewMockLomsService(ctrl)
	srv := NewLomsServer(lomsService, zap.NewNop())

	req := &lomsv1.CreateOrderRequest{
		UserId: 1,
		Items: []*lomsv1.Item{
			{
				Sku:   100,
				Count: 2,
			},
			{
				Sku:   200,
				Count: 1,
			},
		},
	}

	expectedItems := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	lomsService.EXPECT().
		CreateOrder(gomock.Any(), req.UserId, expectedItems).
		Return(int64(123), nil)

	resp, err := srv.CreateOrder(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int64(123), resp.OrderId)
}

func TestLomsServer_CreateOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
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

			lomsService := mocks.NewMockLomsService(ctrl)
			srv := NewLomsServer(lomsService, zap.NewNop())

			req := &lomsv1.CreateOrderRequest{
				UserId: 1,
				Items: []*lomsv1.Item{
					{
						Sku:   100,
						Count: 2,
					},
				},
			}

			expectedItems := []entity.OrderItem{
				{
					SKU:   100,
					Count: 2,
				},
			}

			lomsService.EXPECT().
				CreateOrder(gomock.Any(), req.UserId, expectedItems).
				Return(int64(0), tt.serviceError)

			resp, err := srv.CreateOrder(context.Background(), req)

			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestLomsServer_GetOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	lomsService := mocks.NewMockLomsService(ctrl)
	srv := NewLomsServer(lomsService, zap.NewNop())

	req := &lomsv1.GetOrderRequest{
		OrderId: 1,
	}

	now := time.Now()
	order := entity.Order{
		UserID: 42,
		Status: entity.OrderStatusNew,
		Items: []entity.OrderItem{
			{
				SKU:   100,
				Count: 2,
			},
			{
				SKU:   200,
				Count: 1,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	lomsService.EXPECT().
		GetOrder(gomock.Any(), req.OrderId).
		Return(order, nil)

	resp, err := srv.GetOrder(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Equal(t, order.UserID, resp.UserId)
	require.Len(t, resp.Items, 2)
	require.Equal(t, now.Unix(), resp.CreatedAt.AsTime().Unix())
	require.Equal(t, now.Unix(), resp.UpdatedAt.AsTime().Unix())
}

func TestLomsServer_GetOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "order not found",
			serviceError: entity.ErrOrderNotFound,
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

			lomsService := mocks.NewMockLomsService(ctrl)
			srv := NewLomsServer(lomsService, zap.NewNop())

			req := &lomsv1.GetOrderRequest{
				OrderId: 1,
			}

			lomsService.EXPECT().
				GetOrder(gomock.Any(), req.OrderId).
				Return(entity.Order{}, tt.serviceError)

			resp, err := srv.GetOrder(context.Background(), req)

			require.Nil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}

func TestLomsServer_PayOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	lomsService := mocks.NewMockLomsService(ctrl)
	srv := NewLomsServer(lomsService, zap.NewNop())

	req := &lomsv1.PayOrderRequest{
		OrderId: 1,
	}

	lomsService.EXPECT().
		PayOrder(gomock.Any(), req.OrderId).
		Return(nil)

	resp, err := srv.PayOrder(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestLomsServer_PayOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceError error
		expectedCode codes.Code
	}{
		{
			name:         "order not found",
			serviceError: entity.ErrOrderNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "order already paid",
			serviceError: entity.ErrOrderAlreadyPaid,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "order cancelled",
			serviceError: entity.ErrOrderCancelled,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "order failed",
			serviceError: entity.ErrOrderFailed,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "unknown order status",
			serviceError: entity.ErrUnknownOrderStatus,
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

			lomsService := mocks.NewMockLomsService(ctrl)
			srv := NewLomsServer(lomsService, zap.NewNop())

			req := &lomsv1.PayOrderRequest{
				OrderId: 1,
			}

			lomsService.EXPECT().
				PayOrder(gomock.Any(), req.OrderId).
				Return(tt.serviceError)

			resp, err := srv.PayOrder(context.Background(), req)

			require.NotNil(t, resp)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.expectedCode, st.Code())
		})
	}
}
