package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/order/mocks"
	sqlcorder "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/order/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPostgresRepository_CreateOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()

	order := entity.Order{
		UserID: 1,
		Status: entity.OrderStatusAwaitingPayment,
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
	}

	orderID := int64(10)

	gomock.InOrder(
		querier.EXPECT().
			InsertOrder(gomock.Any(), sqlcorder.InsertOrderParams{
				UserID: order.UserID,
				Status: sqlcorder.LomsOrderStatusAwaitingpayment,
			}).
			Return(sqlcorder.LomsOrder{
				ID:     orderID,
				UserID: order.UserID,
				Status: sqlcorder.LomsOrderStatusAwaitingpayment,
			}, nil),

		querier.EXPECT().
			InsertOrderItem(gomock.Any(), sqlcorder.InsertOrderItemParams{
				OrderID: orderID,
				Sku:     100,
				Count:   2,
			}).
			Return(nil),

		querier.EXPECT().
			InsertOrderItem(gomock.Any(), sqlcorder.InsertOrderItemParams{
				OrderID: orderID,
				Sku:     200,
				Count:   1,
			}).
			Return(nil),
	)

	gotOrderID, err := repo.CreateOrder(ctx, order)

	require.NoError(t, err)
	require.Equal(t, orderID, gotOrderID)
}

func TestPostgresRepository_CreateOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	insertOrderErr := errors.New("insert order error")
	insertItemErr := errors.New("insert item error")

	tests := []struct {
		name        string
		order       entity.Order
		insertErr   error
		itemErr     error
		expectedErr error
	}{
		{
			name: "insert order error",
			order: entity.Order{
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
				Items: []entity.OrderItem{
					{
						SKU:   100,
						Count: 2,
					},
				},
			},
			insertErr:   insertOrderErr,
			expectedErr: insertOrderErr,
		},
		{
			name: "insert order item error",
			order: entity.Order{
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
				Items: []entity.OrderItem{
					{
						SKU:   100,
						Count: 2,
					},
				},
			},
			itemErr:     insertItemErr,
			expectedErr: insertItemErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			querier := mocks.NewMockQuerier(ctrl)

			repo := &postgresRepository{
				queries: querier,
			}

			orderID := int64(10)

			querier.EXPECT().
				InsertOrder(gomock.Any(), sqlcorder.InsertOrderParams{
					UserID: tt.order.UserID,
					Status: toSQLCStatus(tt.order.Status),
				}).
				Return(sqlcorder.LomsOrder{
					ID:     orderID,
					UserID: tt.order.UserID,
					Status: toSQLCStatus(tt.order.Status),
				}, tt.insertErr)

			if tt.insertErr == nil {
				querier.EXPECT().
					InsertOrderItem(gomock.Any(), sqlcorder.InsertOrderItemParams{
						OrderID: orderID,
						Sku:     int64(tt.order.Items[0].SKU),
						Count:   int64(tt.order.Items[0].Count),
					}).
					Return(tt.itemErr)
			}

			gotOrderID, err := repo.CreateOrder(context.Background(), tt.order)

			require.Error(t, err)
			require.EqualValues(t, 0, gotOrderID)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_GetOrderByID_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)
	userID := int64(1)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	gomock.InOrder(
		querier.EXPECT().
			GetOrderByID(gomock.Any(), orderID).
			Return(sqlcorder.LomsOrder{
				ID:     orderID,
				UserID: userID,
				Status: sqlcorder.LomsOrderStatusPaid,
				CreatedAt: pgtype.Timestamptz{
					Time:  createdAt,
					Valid: true,
				},
				UpdatedAt: pgtype.Timestamptz{
					Time:  updatedAt,
					Valid: true,
				},
			}, nil),

		querier.EXPECT().
			ListOrderItemsByOrderID(gomock.Any(), orderID).
			Return([]sqlcorder.ListOrderItemsByOrderIDRow{
				{

					Sku:   100,
					Count: 2,
				},
				{
					Sku:   200,
					Count: 1,
				},
			}, nil),
	)

	gotOrder, err := repo.GetOrderByID(ctx, orderID)

	require.NoError(t, err)
	require.Equal(t, orderID, gotOrder.ID)
	require.Equal(t, userID, gotOrder.UserID)
	require.Equal(t, entity.OrderStatusPaid, gotOrder.Status)
	require.Equal(t, createdAt, gotOrder.CreatedAt)
	require.Equal(t, updatedAt, gotOrder.UpdatedAt)

	require.Len(t, gotOrder.Items, 2)
	require.EqualValues(t, 100, gotOrder.Items[0].SKU)
	require.EqualValues(t, 2, gotOrder.Items[0].Count)
	require.EqualValues(t, 200, gotOrder.Items[1].SKU)
	require.EqualValues(t, 1, gotOrder.Items[1].Count)
}

func TestPostgresRepository_GetOrderByID_Err_Gomock(t *testing.T) {
	t.Parallel()

	getOrderErr := errors.New("get order error")
	listItemsErr := errors.New("list items error")

	tests := []struct {
		name        string
		getErr      error
		listErr     error
		expectedErr error
	}{
		{
			name:        "order not found",
			getErr:      pgx.ErrNoRows,
			expectedErr: entity.ErrOrderNotFound,
		},
		{
			name:        "get order error",
			getErr:      getOrderErr,
			expectedErr: getOrderErr,
		},
		{
			name:        "list items error",
			listErr:     listItemsErr,
			expectedErr: listItemsErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			querier := mocks.NewMockQuerier(ctrl)

			repo := &postgresRepository{
				queries: querier,
			}

			orderID := int64(10)

			querier.EXPECT().
				GetOrderByID(gomock.Any(), orderID).
				Return(sqlcorder.LomsOrder{
					ID:     orderID,
					UserID: 1,
					Status: sqlcorder.LomsOrderStatusPaid,
				}, tt.getErr)

			if tt.getErr == nil {
				querier.EXPECT().
					ListOrderItemsByOrderID(gomock.Any(), orderID).
					Return(nil, tt.listErr)
			}

			gotOrder, err := repo.GetOrderByID(context.Background(), orderID)

			require.Error(t, err)
			require.Equal(t, entity.Order{}, gotOrder)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_GetOrderFOrUpdateByID_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)
	userID := int64(1)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	gomock.InOrder(
		querier.EXPECT().
			GetOrderForUpdateByID(gomock.Any(), orderID).
			Return(sqlcorder.LomsOrder{
				ID:     orderID,
				UserID: userID,
				Status: sqlcorder.LomsOrderStatusAwaitingpayment,
				CreatedAt: pgtype.Timestamptz{
					Time:  createdAt,
					Valid: true,
				},
				UpdatedAt: pgtype.Timestamptz{
					Time:  updatedAt,
					Valid: true,
				},
			}, nil),

		querier.EXPECT().
			ListOrderItemsByOrderID(gomock.Any(), orderID).
			Return([]sqlcorder.ListOrderItemsByOrderIDRow{
				{
					Sku:   100,
					Count: 2,
				},
			}, nil),
	)

	gotOrder, err := repo.GetOrderFOrUpdateByID(ctx, orderID)

	require.NoError(t, err)
	require.Equal(t, orderID, gotOrder.ID)
	require.Equal(t, userID, gotOrder.UserID)
	require.Equal(t, entity.OrderStatusAwaitingPayment, gotOrder.Status)
	require.Equal(t, createdAt, gotOrder.CreatedAt)
	require.Equal(t, updatedAt, gotOrder.UpdatedAt)

	require.Len(t, gotOrder.Items, 1)
	require.EqualValues(t, 100, gotOrder.Items[0].SKU)
	require.EqualValues(t, 2, gotOrder.Items[0].Count)
}

func TestPostgresRepository_GetOrderFOrUpdateByID_Err_Gomock(t *testing.T) {
	t.Parallel()

	getOrderErr := errors.New("get order for update error")
	listItemsErr := errors.New("list items error")

	tests := []struct {
		name        string
		getErr      error
		listErr     error
		expectedErr error
	}{
		{
			name:        "order not found",
			getErr:      pgx.ErrNoRows,
			expectedErr: entity.ErrOrderNotFound,
		},
		{
			name:        "get order error",
			getErr:      getOrderErr,
			expectedErr: getOrderErr,
		},
		{
			name:        "list items error",
			listErr:     listItemsErr,
			expectedErr: listItemsErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			querier := mocks.NewMockQuerier(ctrl)

			repo := &postgresRepository{
				queries: querier,
			}

			orderID := int64(10)

			querier.EXPECT().
				GetOrderForUpdateByID(gomock.Any(), orderID).
				Return(sqlcorder.LomsOrder{
					ID:     orderID,
					UserID: 1,
					Status: sqlcorder.LomsOrderStatusPaid,
				}, tt.getErr)

			if tt.getErr == nil {
				querier.EXPECT().
					ListOrderItemsByOrderID(gomock.Any(), orderID).
					Return(nil, tt.listErr)
			}

			gotOrder, err := repo.GetOrderFOrUpdateByID(context.Background(), orderID)

			require.Error(t, err)
			require.Equal(t, entity.Order{}, gotOrder)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_SetStatusByID_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)
	status := entity.OrderStatusPaid

	querier.EXPECT().
		SetOrderStatus(gomock.Any(), sqlcorder.SetOrderStatusParams{
			ID:     orderID,
			Status: sqlcorder.LomsOrderStatusPaid,
		}).
		Return(nil)

	err := repo.SetStatusByID(ctx, orderID, status)

	require.NoError(t, err)
}

func TestPostgresRepository_SetStatusByID_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)
	status := entity.OrderStatusCancelled
	expectedErr := errors.New("set status error")

	querier.EXPECT().
		SetOrderStatus(gomock.Any(), sqlcorder.SetOrderStatusParams{
			ID:     orderID,
			Status: sqlcorder.LomsOrderStatusCancelled,
		}).
		Return(expectedErr)

	err := repo.SetStatusByID(ctx, orderID, status)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestToSQLCStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   entity.OrderStatus
		expected sqlcorder.LomsOrderStatus
	}{
		{
			name:     "new",
			status:   entity.OrderStatusNew,
			expected: sqlcorder.LomsOrderStatusNew,
		},
		{
			name:     "awaiting payment",
			status:   entity.OrderStatusAwaitingPayment,
			expected: sqlcorder.LomsOrderStatusAwaitingpayment,
		},
		{
			name:     "failed",
			status:   entity.OrderStatusFailed,
			expected: sqlcorder.LomsOrderStatusFailed,
		},
		{
			name:     "paid",
			status:   entity.OrderStatusPaid,
			expected: sqlcorder.LomsOrderStatusPaid,
		},
		{
			name:     "cancelled",
			status:   entity.OrderStatusCancelled,
			expected: sqlcorder.LomsOrderStatusCancelled,
		},
		{
			name:     "unknown fallback",
			status:   entity.OrderStatusUnknown,
			expected: sqlcorder.LomsOrderStatusNew,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toSQLCStatus(tt.status)

			require.Equal(t, tt.expected, got)
		})
	}
}

func TestToEntityStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   sqlcorder.LomsOrderStatus
		expected entity.OrderStatus
	}{
		{
			name:     "new",
			status:   sqlcorder.LomsOrderStatusNew,
			expected: entity.OrderStatusNew,
		},
		{
			name:     "awaiting payment",
			status:   sqlcorder.LomsOrderStatusAwaitingpayment,
			expected: entity.OrderStatusAwaitingPayment,
		},
		{
			name:     "failed",
			status:   sqlcorder.LomsOrderStatusFailed,
			expected: entity.OrderStatusFailed,
		},
		{
			name:     "paid",
			status:   sqlcorder.LomsOrderStatusPaid,
			expected: entity.OrderStatusPaid,
		},
		{
			name:     "cancelled",
			status:   sqlcorder.LomsOrderStatusCancelled,
			expected: entity.OrderStatusCancelled,
		},
		{
			name:     "unknown fallback",
			status:   sqlcorder.LomsOrderStatus("some_unknown_status"),
			expected: entity.OrderStatusNew,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toEntityStatus(tt.status)

			require.Equal(t, tt.expected, got)
		})
	}
}
