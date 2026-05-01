package notifications

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/controller/notifications/mocks"
	notificationsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/notifications/api/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNotificationsServer_SendMessage_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	notificationsService := mocks.NewMockService(ctrl)

	srv := NewNotificationsServer(notificationsService, zap.NewNop())

	req := &notificationsv1.SendMessageRequest{
		UserId:  1,
		OrderId: 10,
		Status:  notificationsv1.OrderStatus_ORDER_STATUS_PAID,
	}

	notificationsService.EXPECT().
		SendMessage(
			gomock.Any(),
			req.GetUserId(),
			req.GetOrderId(),
			req.GetStatus().String(),
		).
		Return(nil)

	// Act
	resp, err := srv.SendMessage(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestNotificationsServer_SendMessage_ServiceErr_Gomock(t *testing.T) {
	t.Parallel()

	serviceErr := errors.New("send message error")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	notificationsService := mocks.NewMockService(ctrl)

	srv := NewNotificationsServer(notificationsService, zap.NewNop())

	req := &notificationsv1.SendMessageRequest{
		UserId:  1,
		OrderId: 10,
		Status:  notificationsv1.OrderStatus_ORDER_STATUS_PAID,
	}

	notificationsService.EXPECT().
		SendMessage(
			gomock.Any(),
			req.GetUserId(),
			req.GetOrderId(),
			req.GetStatus().String(),
		).
		Return(serviceErr)

	// Act
	resp, err := srv.SendMessage(context.Background(), req)

	// Assert
	require.Nil(t, resp)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, st.Code())
	require.Contains(t, st.Message(), serviceErr.Error())
}



