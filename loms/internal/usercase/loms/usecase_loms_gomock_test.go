package loms

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/adapter/notifications/converter"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/port"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/usercase/loms/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLomsService_CreateOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()
	userID := int64(1)
	orderID := int64(10)

	items := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	gomock.InOrder(
		orderRepository.EXPECT().
			CreateOrder(gomock.Any(), gomock.AssignableToTypeOf(entity.Order{})).
			DoAndReturn(func(ctx context.Context, order entity.Order) (int64, error) {
				require.Equal(t, userID, order.UserID)
				require.Equal(t, items, order.Items)
				require.Equal(t, entity.OrderStatusAwaitingPayment, order.Status)
				require.False(t, order.CreatedAt.IsZero())
				require.False(t, order.UpdatedAt.IsZero())

				return orderID, nil
			}),

		stocksRepository.EXPECT().
			ReserveStocks(gomock.Any(), orderID, items).
			Return(nil),

		outboxRepository.EXPECT().
			SendMessage(
				gomock.Any(),
				srv.createKey(orderID, entity.OrderStatusAwaitingPayment),
				outbox.KindNotification,
				gomock.Any(),
			).
			Return(nil),
	)

	gotOrderID, err := srv.CreateOrder(ctx, userID, items)

	require.NoError(t, err)
	require.Equal(t, orderID, gotOrderID)
}

func TestLomsService_CreateOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	txErr := errors.New("tx error")
	createOrderErr := errors.New("create order error")
	reserveStocksErr := errors.New("reserve stocks error")
	outboxErr := errors.New("outbox error")

	tests := []struct {
		name        string
		txErr       error
		createErr   error
		reserveErr  error
		outboxErr   error
		expectedErr error
	}{
		{
			name:        "transactor error",
			txErr:       txErr,
			expectedErr: txErr,
		},
		{
			name:        "create order error",
			createErr:   createOrderErr,
			expectedErr: createOrderErr,
		},
		{
			name:        "reserve stocks error",
			reserveErr:  reserveStocksErr,
			expectedErr: reserveStocksErr,
		},
		{
			name:        "outbox error",
			outboxErr:   outboxErr,
			expectedErr: outboxErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			orderRepository := mocks.NewMockorderRepository(ctrl)
			stocksRepository := mocks.NewMockstocksRepository(ctrl)
			transactor := mocks.NewMocktransactor(ctrl)
			outboxRepository := mocks.NewMockoutboxRepository(ctrl)
			notificationsClient := mocks.NewMocknotificationsClient(ctrl)

			srv := NewLomsService(
				orderRepository,
				stocksRepository,
				transactor,
				notificationsClient,
				outboxRepository,
			)

			ctx := context.Background()
			userID := int64(1)
			orderID := int64(10)

			items := []entity.OrderItem{
				{
					SKU:   100,
					Count: 2,
				},
			}

			if tt.txErr != nil {
				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					Return(tt.txErr)
			} else {
				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})

				orderRepository.EXPECT().
					CreateOrder(gomock.Any(), gomock.AssignableToTypeOf(entity.Order{})).
					Return(orderID, tt.createErr)

				if tt.createErr == nil {
					stocksRepository.EXPECT().
						ReserveStocks(gomock.Any(), orderID, items).
						Return(tt.reserveErr)
				}

				if tt.createErr == nil && tt.reserveErr == nil {
					outboxRepository.EXPECT().
						SendMessage(
							gomock.Any(),
							srv.createKey(orderID, entity.OrderStatusAwaitingPayment),
							outbox.KindNotification,
							gomock.Any(),
						).
						Return(tt.outboxErr)
				}
			}

			gotOrderID, err := srv.CreateOrder(ctx, userID, items)

			require.Error(t, err)
			require.EqualValues(t, 0, gotOrderID)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestLomsService_GetOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()
	orderID := int64(10)

	expectedOrder := entity.Order{
		ID:        orderID,
		UserID:    1,
		Status:    entity.OrderStatusAwaitingPayment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Items: []entity.OrderItem{
			{
				SKU:   100,
				Count: 2,
			},
		},
	}

	orderRepository.EXPECT().
		GetOrderByID(gomock.Any(), orderID).
		Return(expectedOrder, nil)

	order, err := srv.GetOrder(ctx, orderID)

	require.NoError(t, err)
	require.Equal(t, expectedOrder, order)
}

func TestLomsService_GetOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()
	orderID := int64(10)
	expectedErr := errors.New("get order error")

	orderRepository.EXPECT().
		GetOrderByID(gomock.Any(), orderID).
		Return(entity.Order{}, expectedErr)

	order, err := srv.GetOrder(ctx, orderID)

	require.Error(t, err)
	require.Equal(t, entity.Order{}, order)
	require.ErrorIs(t, err, expectedErr)
}

func TestLomsService_PayOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()
	orderID := int64(10)

	order := entity.Order{
		ID:     orderID,
		UserID: 1,
		Status: entity.OrderStatusAwaitingPayment,
		Items: []entity.OrderItem{
			{
				SKU:   100,
				Count: 2,
			},
		},
	}

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	gomock.InOrder(
		orderRepository.EXPECT().
			GetOrderFOrUpdateByID(gomock.Any(), orderID).
			Return(order, nil),

		orderRepository.EXPECT().
			SetStatusByID(gomock.Any(), orderID, entity.OrderStatusPaid).
			Return(nil),

		outboxRepository.EXPECT().
			SendMessage(
				gomock.Any(),
				srv.createKey(orderID, entity.OrderStatusPaid),
				outbox.KindNotification,
				gomock.Any(),
			).
			Return(nil),
	)

	err := srv.PayOrder(ctx, orderID)

	require.NoError(t, err)
}

func TestLomsService_PayOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	txErr := errors.New("tx error")
	getOrderErr := errors.New("get order error")
	setStatusErr := errors.New("set status error")
	outboxErr := errors.New("outbox error")

	tests := []struct {
		name        string
		txErr       error
		order       entity.Order
		getErr      error
		setErr      error
		outboxErr   error
		expectedErr error
	}{
		{
			name:        "transactor error",
			txErr:       txErr,
			expectedErr: txErr,
		},
		{
			name:        "get order error",
			getErr:      getOrderErr,
			expectedErr: getOrderErr,
		},
		{
			name: "order cancelled",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusCancelled,
			},
			expectedErr: entity.ErrOrderCancelled,
		},
		{
			name: "order already paid",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusPaid,
			},
			expectedErr: entity.ErrOrderAlreadyPaid,
		},
		{
			name: "order failed",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusFailed,
			},
			expectedErr: entity.ErrOrderFailed,
		},
		{
			name: "unknown status",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusUnknown,
			},
			expectedErr: entity.ErrUnknownOrderStatus,
		},
		{
			name: "set status error",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
			},
			setErr:      setStatusErr,
			expectedErr: setStatusErr,
		},
		{
			name: "outbox error",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
			},
			outboxErr:   outboxErr,
			expectedErr: outboxErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			orderRepository := mocks.NewMockorderRepository(ctrl)
			stocksRepository := mocks.NewMockstocksRepository(ctrl)
			transactor := mocks.NewMocktransactor(ctrl)
			outboxRepository := mocks.NewMockoutboxRepository(ctrl)
			notificationsClient := mocks.NewMocknotificationsClient(ctrl)

			srv := NewLomsService(
				orderRepository,
				stocksRepository,
				transactor,
				notificationsClient,
				outboxRepository,
			)

			ctx := context.Background()
			orderID := int64(10)

			if tt.txErr != nil {
				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					Return(tt.txErr)
			} else {
				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})

				orderRepository.EXPECT().
					GetOrderFOrUpdateByID(gomock.Any(), orderID).
					Return(tt.order, tt.getErr)

				if tt.getErr == nil &&
					(tt.order.Status == entity.OrderStatusNew ||
						tt.order.Status == entity.OrderStatusAwaitingPayment) {
					orderRepository.EXPECT().
						SetStatusByID(gomock.Any(), orderID, entity.OrderStatusPaid).
						Return(tt.setErr)

					if tt.setErr == nil {
						outboxRepository.EXPECT().
							SendMessage(
								gomock.Any(),
								srv.createKey(tt.order.ID, entity.OrderStatusPaid),
								outbox.KindNotification,
								gomock.Any(),
							).
							Return(tt.outboxErr)
					}
				}
			}

			err := srv.PayOrder(ctx, orderID)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestLomsService_CancelOrder_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()
	orderID := int64(10)

	items := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
	}

	order := entity.Order{
		ID:     orderID,
		UserID: 1,
		Status: entity.OrderStatusAwaitingPayment,
		Items:  items,
	}

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	gomock.InOrder(
		orderRepository.EXPECT().
			GetOrderFOrUpdateByID(gomock.Any(), orderID).
			Return(order, nil),

		stocksRepository.EXPECT().
			ReleaseStocks(gomock.Any(), orderID, items).
			Return(nil),

		orderRepository.EXPECT().
			SetStatusByID(gomock.Any(), orderID, entity.OrderStatusCancelled).
			Return(nil),

		outboxRepository.EXPECT().
			SendMessage(
				gomock.Any(),
				srv.createKey(orderID, entity.OrderStatusCancelled),
				outbox.KindNotification,
				gomock.Any(),
			).
			Return(nil),
	)

	err := srv.CancelOrder(ctx, orderID)

	require.NoError(t, err)
}

func TestLomsService_CancelOrder_Err_Gomock(t *testing.T) {
	t.Parallel()

	txErr := errors.New("tx error")
	getOrderErr := errors.New("get order error")
	releaseErr := errors.New("release stocks error")
	setStatusErr := errors.New("set status error")
	outboxErr := errors.New("outbox error")

	tests := []struct {
		name        string
		txErr       error
		order       entity.Order
		getErr      error
		releaseErr  error
		setErr      error
		outboxErr   error
		expectedErr error
	}{
		{
			name:        "transactor error",
			txErr:       txErr,
			expectedErr: txErr,
		},
		{
			name:        "get order error",
			getErr:      getOrderErr,
			expectedErr: getOrderErr,
		},
		{
			name: "order cancelled",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusCancelled,
			},
			expectedErr: entity.ErrOrderCancelled,
		},
		{
			name: "order already paid",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusPaid,
			},
			expectedErr: entity.ErrOrderAlreadyPaid,
		},
		{
			name: "release stocks error",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
				Items: []entity.OrderItem{
					{
						SKU:   100,
						Count: 2,
					},
				},
			},
			releaseErr:  releaseErr,
			expectedErr: releaseErr,
		},
		{
			name: "set status error",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
				Items: []entity.OrderItem{
					{
						SKU:   100,
						Count: 2,
					},
				},
			},
			setErr:      setStatusErr,
			expectedErr: setStatusErr,
		},
		{
			name: "outbox error",
			order: entity.Order{
				ID:     10,
				UserID: 1,
				Status: entity.OrderStatusAwaitingPayment,
				Items: []entity.OrderItem{
					{
						SKU:   100,
						Count: 2,
					},
				},
			},
			outboxErr:   outboxErr,
			expectedErr: outboxErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			orderRepository := mocks.NewMockorderRepository(ctrl)
			stocksRepository := mocks.NewMockstocksRepository(ctrl)
			transactor := mocks.NewMocktransactor(ctrl)
			outboxRepository := mocks.NewMockoutboxRepository(ctrl)
			notificationsClient := mocks.NewMocknotificationsClient(ctrl)

			srv := NewLomsService(
				orderRepository,
				stocksRepository,
				transactor,
				notificationsClient,
				outboxRepository,
			)

			ctx := context.Background()
			orderID := int64(10)

			if tt.txErr != nil {
				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					Return(tt.txErr)
			} else {
				transactor.EXPECT().
					WithTx(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})

				orderRepository.EXPECT().
					GetOrderFOrUpdateByID(gomock.Any(), orderID).
					Return(tt.order, tt.getErr)

				if tt.getErr == nil &&
					tt.order.Status != entity.OrderStatusCancelled &&
					tt.order.Status != entity.OrderStatusPaid {
					stocksRepository.EXPECT().
						ReleaseStocks(gomock.Any(), orderID, tt.order.Items).
						Return(tt.releaseErr)

					if tt.releaseErr == nil {
						orderRepository.EXPECT().
							SetStatusByID(gomock.Any(), orderID, entity.OrderStatusCancelled).
							Return(tt.setErr)
					}

					if tt.releaseErr == nil && tt.setErr == nil {
						outboxRepository.EXPECT().
							SendMessage(
								gomock.Any(),
								srv.createKey(tt.order.ID, entity.OrderStatusCancelled),
								outbox.KindNotification,
								gomock.Any(),
							).
							Return(tt.outboxErr)
					}
				}
			}

			err := srv.CancelOrder(ctx, orderID)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestLomsService_OrderStatusChangedNotificationKindHandler_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()

	body := port.OrderStatusChangedNotification{
		UserID:  1,
		OrderID: 10,
		Status:  converter.FromOrderStatus(entity.OrderStatusPaid),
	}

	data, err := json.Marshal(body)
	require.NoError(t, err)

	notificationsClient.EXPECT().
		SendMessage(gomock.Any(), body.UserID, body.OrderID, body.Status).
		Return(nil)

	err = srv.OrderStatusChangedNotificationKindHandler(ctx, data)

	require.NoError(t, err)
}

func TestLomsService_OrderStatusChangedNotificationKindHandler_InvalidJSON_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	err := srv.OrderStatusChangedNotificationKindHandler(context.Background(), []byte("{"))

	require.Error(t, err)
}

func TestLomsService_OrderStatusChangedNotificationKindHandler_SendMessageErr_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	orderRepository := mocks.NewMockorderRepository(ctrl)
	stocksRepository := mocks.NewMockstocksRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	outboxRepository := mocks.NewMockoutboxRepository(ctrl)
	notificationsClient := mocks.NewMocknotificationsClient(ctrl)

	srv := NewLomsService(
		orderRepository,
		stocksRepository,
		transactor,
		notificationsClient,
		outboxRepository,
	)

	ctx := context.Background()
	expectedErr := errors.New("send notification error")

	body := port.OrderStatusChangedNotification{
		UserID:  1,
		OrderID: 10,
		Status:  converter.FromOrderStatus(entity.OrderStatusPaid),
	}

	data, err := json.Marshal(body)
	require.NoError(t, err)

	notificationsClient.EXPECT().
		SendMessage(gomock.Any(), body.UserID, body.OrderID, body.Status).
		Return(expectedErr)

	err = srv.OrderStatusChangedNotificationKindHandler(ctx, data)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestLomsService_CreateKey(t *testing.T) {
	t.Parallel()

	srv := &lomsService{}

	key := srv.createKey(10, entity.OrderStatusPaid)

	require.Equal(t, "10_"+string(entity.OrderStatusPaid), key)
}



