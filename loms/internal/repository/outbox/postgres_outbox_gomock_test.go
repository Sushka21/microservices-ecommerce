package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox/mocks"
	sqlc "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOutboxRepository_SendMessage_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	idempotencyKey := "10_paid"
	kind := KindNotification
	message := []byte(`{"order_id":10,"status":"paid"}`)

	querier.EXPECT().
		SendOutboxMessage(gomock.Any(), sqlc.SendOutboxMessageParams{
			IdempotencyKey: idempotencyKey,
			Data:           message,
			Kind:           int32(kind),
		}).
		Return(nil)

	err := repo.SendMessage(ctx, idempotencyKey, kind, message)

	require.NoError(t, err)
}

func TestOutboxRepository_SendMessage_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	idempotencyKey := "10_paid"
	kind := KindNotification
	message := []byte(`{"order_id":10,"status":"paid"}`)
	expectedErr := errors.New("send outbox message error")

	querier.EXPECT().
		SendOutboxMessage(gomock.Any(), sqlc.SendOutboxMessageParams{
			IdempotencyKey: idempotencyKey,
			Data:           message,
			Kind:           int32(kind),
		}).
		Return(expectedErr)

	err := repo.SendMessage(ctx, idempotencyKey, kind, message)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestOutboxRepository_GetMessages_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	batchSize := 10
	inProgressTTL := 30 * time.Second

	rows := []sqlc.GetOutboxMessagesRow{
		{
			IdempotencyKey: "10_paid",
			Data:           []byte(`{"order_id":10,"status":"paid"}`),
			Kind:           int32(KindNotification),
		},
		{
			IdempotencyKey: "11_cancelled",
			Data:           []byte(`{"order_id":11,"status":"cancelled"}`),
			Kind:           int32(KindNotification),
		},
	}

	querier.EXPECT().
		GetOutboxMessages(gomock.Any(), sqlc.GetOutboxMessagesParams{
			InProgressTtl: pgtype.Interval{
				Microseconds: inProgressTTL.Microseconds(),
				Valid:        true,
			},
			BatchSize: int32(batchSize),
		}).
		Return(rows, nil)

	messages, err := repo.GetMessages(ctx, batchSize, inProgressTTL)

	require.NoError(t, err)
	require.Len(t, messages, 2)

	require.Equal(t, "10_paid", messages[0].IdempotencyKey)
	require.Equal(t, []byte(`{"order_id":10,"status":"paid"}`), messages[0].Data)
	require.Equal(t, KindNotification, messages[0].Kind)

	require.Equal(t, "11_cancelled", messages[1].IdempotencyKey)
	require.Equal(t, []byte(`{"order_id":11,"status":"cancelled"}`), messages[1].Data)
	require.Equal(t, KindNotification, messages[1].Kind)
}

func TestOutboxRepository_GetMessages_Empty_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	batchSize := 10
	inProgressTTL := 30 * time.Second

	querier.EXPECT().
		GetOutboxMessages(gomock.Any(), sqlc.GetOutboxMessagesParams{
			InProgressTtl: pgtype.Interval{
				Microseconds: inProgressTTL.Microseconds(),
				Valid:        true,
			},
			BatchSize: int32(batchSize),
		}).
		Return([]sqlc.GetOutboxMessagesRow{}, nil)

	messages, err := repo.GetMessages(ctx, batchSize, inProgressTTL)

	require.NoError(t, err)
	require.Empty(t, messages)
}

func TestOutboxRepository_GetMessages_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	batchSize := 10
	inProgressTTL := 30 * time.Second
	expectedErr := errors.New("get outbox messages error")

	querier.EXPECT().
		GetOutboxMessages(gomock.Any(), sqlc.GetOutboxMessagesParams{
			InProgressTtl: pgtype.Interval{
				Microseconds: inProgressTTL.Microseconds(),
				Valid:        true,
			},
			BatchSize: int32(batchSize),
		}).
		Return(nil, expectedErr)

	messages, err := repo.GetMessages(ctx, batchSize, inProgressTTL)

	require.Error(t, err)
	require.Nil(t, messages)
	require.ErrorIs(t, err, expectedErr)
}

func TestOutboxRepository_MarkAsProcessed_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	keys := []string{"10_paid", "11_cancelled"}

	querier.EXPECT().
		MarkOutboxMessagesAsProcessed(gomock.Any(), keys).
		Return(nil)

	err := repo.MarkAsProcessed(ctx, keys)

	require.NoError(t, err)
}

func TestOutboxRepository_MarkAsProcessed_EmptyKeys_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	err := repo.MarkAsProcessed(context.Background(), nil)

	require.NoError(t, err)
}

func TestOutboxRepository_MarkAsProcessed_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	keys := []string{"10_paid"}
	expectedErr := errors.New("mark processed error")

	querier.EXPECT().
		MarkOutboxMessagesAsProcessed(gomock.Any(), keys).
		Return(expectedErr)

	err := repo.MarkAsProcessed(ctx, keys)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestOutboxRepository_MarkAsRetryable_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	keys := []string{"10_paid", "11_cancelled"}

	querier.EXPECT().
		MarkOutboxMessagesAsRetryable(gomock.Any(), keys).
		Return(nil)

	err := repo.MarkAsRetryable(ctx, keys)

	require.NoError(t, err)
}

func TestOutboxRepository_MarkAsRetryable_EmptyKeys_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	err := repo.MarkAsRetryable(context.Background(), []string{})

	require.NoError(t, err)
}

func TestOutboxRepository_MarkAsRetryable_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &outboxRepository{
		queries: querier,
	}

	ctx := context.Background()
	keys := []string{"10_paid"}
	expectedErr := errors.New("mark retryable error")

	querier.EXPECT().
		MarkOutboxMessagesAsRetryable(gomock.Any(), keys).
		Return(expectedErr)

	err := repo.MarkAsRetryable(ctx, keys)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}
