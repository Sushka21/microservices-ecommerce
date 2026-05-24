package inbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/inbox/mocks"
	repository "github.com/Sushka21/microservices-ecommerce/notifications/internal/repository/inbox"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	logger := zap.NewNop()
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)
	cfg := &config.Config{}

	handler := func(kind repository.Kind) (KindHandler, error) {
		return func(ctx context.Context, data []byte) error {
			return nil
		}, nil
	}

	// Act
	got := New(logger, repo, handler, cfg, transactor)

	// Assert
	require.NotNil(t, got)
	require.Equal(t, logger, got.logger)
	require.Equal(t, repo, got.outboxRepository)
	require.Equal(t, cfg, got.cfg)
	require.Equal(t, transactor, got.transactor)
	require.NotNil(t, got.globalHandler)
}

func TestOutboxImpl_Worker_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	batchSize := 10
	waitTime := time.Millisecond
	inProgressTTL := time.Minute

	messages := []repository.Data{
		{
			IdempotencyKey: "key-1",
			Kind:           repository.KindNotification,
			Data:           []byte(`{"message":"hello"}`),
		},
		{
			IdempotencyKey: "key-2",
			Kind:           repository.KindNotification,
			Data:           []byte(`{"message":"world"}`),
		},
	}

	handled := make([]string, 0, len(messages))

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		require.Equal(t, repository.KindNotification, kind)

		return func(ctx context.Context, data []byte) error {
			handled = append(handled, string(data))
			return nil
		}, nil
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			err := f(ctx)
			cancel()
			return err
		})

	repo.EXPECT().
		GetMessages(gomock.Any(), batchSize, inProgressTTL).
		Return(messages, nil)

	repo.EXPECT().
		MarkAsProcessed(gomock.Any(), []string{"key-1", "key-2"}).
		Return(nil)

	repo.EXPECT().
		MarkAsRetryable(gomock.Any(), []string{}).
		Return(nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)

	// Assert
	require.Len(t, handled, 2)
	require.Contains(t, handled, `{"message":"hello"}`)
	require.Contains(t, handled, `{"message":"world"}`)
}

func TestOutboxImpl_Worker_HandlerError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	batchSize := 10
	waitTime := time.Millisecond
	inProgressTTL := time.Minute

	messages := []repository.Data{
		{
			IdempotencyKey: "success-key",
			Kind:           repository.KindNotification,
			Data:           []byte(`success`),
		},
		{
			IdempotencyKey: "failed-key",
			Kind:           repository.KindNotification,
			Data:           []byte(`fail`),
		},
	}

	handlerErr := errors.New("handler failed")

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		return func(ctx context.Context, data []byte) error {
			if string(data) == "fail" {
				return handlerErr
			}

			return nil
		}, nil
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			err := f(ctx)
			cancel()
			return err
		})

	repo.EXPECT().
		GetMessages(gomock.Any(), batchSize, inProgressTTL).
		Return(messages, nil)

	repo.EXPECT().
		MarkAsProcessed(gomock.Any(), []string{"success-key"}).
		Return(nil)

	repo.EXPECT().
		MarkAsRetryable(gomock.Any(), []string{"failed-key"}).
		Return(nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)

	// Assert
	require.ErrorIs(t, handlerErr, handlerErr)
}

func TestOutboxImpl_Worker_GlobalHandlerError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	batchSize := 10
	waitTime := time.Millisecond
	inProgressTTL := time.Minute

	messages := []repository.Data{
		{
			IdempotencyKey: "key-1",
			Kind:           repository.KindNotification,
			Data:           []byte(`data`),
		},
	}

	globalHandlerErr := errors.New("unknown kind")

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		return nil, globalHandlerErr
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			err := f(ctx)
			cancel()
			return err
		})

	repo.EXPECT().
		GetMessages(gomock.Any(), batchSize, inProgressTTL).
		Return(messages, nil)

	repo.EXPECT().
		MarkAsProcessed(gomock.Any(), []string{}).
		Return(nil)

	repo.EXPECT().
		MarkAsRetryable(gomock.Any(), []string{}).
		Return(nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)

	// Assert
	require.ErrorIs(t, globalHandlerErr, globalHandlerErr)
}

func TestOutboxImpl_Worker_GetMessagesError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	batchSize := 10
	waitTime := time.Millisecond
	inProgressTTL := time.Minute

	expectedErr := errors.New("get messages failed")

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		return func(ctx context.Context, data []byte) error {
			return nil
		}, nil
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			err := f(ctx)
			cancel()
			return err
		})

	repo.EXPECT().
		GetMessages(gomock.Any(), batchSize, inProgressTTL).
		Return(nil, expectedErr)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)

	// Assert
	require.ErrorIs(t, expectedErr, expectedErr)
}

func TestOutboxImpl_Worker_MarkAsProcessedError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	batchSize := 10
	waitTime := time.Millisecond
	inProgressTTL := time.Minute

	expectedErr := errors.New("mark as processed failed")

	messages := []repository.Data{
		{
			IdempotencyKey: "key-1",
			Kind:           repository.KindNotification,
			Data:           []byte(`data`),
		},
	}

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		return func(ctx context.Context, data []byte) error {
			return nil
		}, nil
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			err := f(ctx)
			cancel()
			return err
		})

	repo.EXPECT().
		GetMessages(gomock.Any(), batchSize, inProgressTTL).
		Return(messages, nil)

	repo.EXPECT().
		MarkAsProcessed(gomock.Any(), []string{"key-1"}).
		Return(expectedErr)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)

	// Assert
	require.ErrorIs(t, expectedErr, expectedErr)
}

func TestOutboxImpl_Worker_MarkAsRetryableError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	batchSize := 10
	waitTime := time.Millisecond
	inProgressTTL := time.Minute

	expectedErr := errors.New("mark as retryable failed")

	messages := []repository.Data{
		{
			IdempotencyKey: "key-1",
			Kind:           repository.KindNotification,
			Data:           []byte(`data`),
		},
	}

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		return func(ctx context.Context, data []byte) error {
			return expectedErr
		}, nil
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			err := f(ctx)
			cancel()
			return err
		})

	repo.EXPECT().
		GetMessages(gomock.Any(), batchSize, inProgressTTL).
		Return(messages, nil)

	repo.EXPECT().
		MarkAsProcessed(gomock.Any(), []string{}).
		Return(nil)

	repo.EXPECT().
		MarkAsRetryable(gomock.Any(), []string{"key-1"}).
		Return(expectedErr)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, batchSize, waitTime, inProgressTTL)

	// Assert
	require.ErrorIs(t, expectedErr, expectedErr)
}

func TestOutboxImpl_Worker_ContextCanceled_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	repo := mocks.NewMockinboxRepository(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	globalHandler := func(kind repository.Kind) (KindHandler, error) {
		return func(ctx context.Context, data []byte) error {
			return nil
		}, nil
	}

	o := New(
		zap.NewNop(),
		repo,
		globalHandler,
		&config.Config{},
		transactor,
	)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// Act
	o.worker(ctx, wg, 10, time.Millisecond, time.Minute)

	// Assert
	require.ErrorIs(t, ctx.Err(), context.Canceled)
}
