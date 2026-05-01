package outbox

import (
	"context"
	"sync"
	"time"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/config"
	repository "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox"
	"go.uber.org/zap"
)

type (
	outboxRepository interface {
		GetMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]repository.Data, error)
		MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error
		MarkAsRetryable(ctx context.Context, idempotencyKeys []string) error
	}

	transactor interface {
		WithTx(ctx context.Context, f func(ctx context.Context) error) (err error)
	}
)

type GlobalHandler = func(kind repository.Kind) (KindHandler, error)
type KindHandler = func(ctx context.Context, data []byte) error

type Outbox interface {
	Start(ctx context.Context, workers int, batchSize int, waitTime time.Duration, inProgressTTL time.Duration)
}

type outboxImpl struct {
	logger           *zap.Logger
	outboxRepository outboxRepository
	globalHandler    GlobalHandler
	cfg              *config.Config
	transactor       transactor
}

func New(
	logger *zap.Logger,
	outboxRepository outboxRepository,
	globalHandler GlobalHandler,
	cfg *config.Config,
	transactor transactor,
) *outboxImpl {
	return &outboxImpl{
		logger:           logger,
		outboxRepository: outboxRepository,
		globalHandler:    globalHandler,
		cfg:              cfg,
		transactor:       transactor,
	}
}

func (o *outboxImpl) Start(
	ctx context.Context,
	workers int,
	batchSize int,
	fetchPeriod time.Duration,
	inProgressTTL time.Duration,
) {
	wg := new(sync.WaitGroup)

	for workerID := 1; workerID <= workers; workerID++ {
		wg.Add(1)
		go o.worker(ctx, wg, batchSize, fetchPeriod, inProgressTTL)
	}
}

func (o *outboxImpl) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	batchSize int,
	waitTime time.Duration,
	inProgressTTL time.Duration,
) {
	defer wg.Done()
	ticker := time.NewTicker(waitTime)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := o.transactor.WithTx(ctx, func(ctx context.Context) error {
				messages, err := o.outboxRepository.GetMessages(ctx, batchSize, inProgressTTL)

				if err != nil {
					o.logger.Error("can not fetch messages from outbox", zap.Error(err))
					return err
				}

				o.logger.Info("messages fetched", zap.Int("size", len(messages)))

				successKeys := make([]string, 0, len(messages)/2)
				failedKeys := make([]string, 0, len(messages)/2)

				for i := range messages {
					message := messages[i]
					key := message.IdempotencyKey

					kindHandler, errKind := o.globalHandler(message.Kind)

					if errKind != nil {
						o.logger.Error("unexpected kind", zap.Error(err))
						continue
					}

					err = kindHandler(ctx, message.Data)

					if err != nil {
						failedKeys = append(failedKeys, key)
						o.logger.Error("kind error", zap.Error(err))
						continue
					}

					successKeys = append(successKeys, key)
				}

				err = o.outboxRepository.MarkAsProcessed(ctx, successKeys)

				if err != nil {
					o.logger.Error("mark as processed outbox error", zap.Error(err))
					return err
				}

				err = o.outboxRepository.MarkAsRetryable(ctx, failedKeys)

				if err != nil {
					o.logger.Error("mark as retryable error", zap.Error(err))
					return err
				}

				return nil
			})

			if err != nil {
				o.logger.Error("worker stage error", zap.Error(err))
			}
		}
	}
}



