package outbox

import (
	"context"
	"fmt"
	"time"

	sqlc "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox/sqlc"
	transactor "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:generate mockgen -source=sqlc/querier.go -destination=mocks/querier_mocks.go -package=mocks

type (
	DB interface {
		sqlc.DBTX
	}
)

type outboxRepository struct {
	db      DB
	queries sqlc.Querier
}

func NewOutboxRepository(db DB) *outboxRepository {
	return &outboxRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (o *outboxRepository) getQueries(ctx context.Context) sqlc.Querier {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return sqlc.New(tx)
	}

	return o.queries
}

func (o *outboxRepository) SendMessage(
	ctx context.Context,
	idempotencyKey string,
	kind Kind,
	message []byte,
) error {
	queries := o.getQueries(ctx)
	err := queries.SendOutboxMessage(ctx, sqlc.SendOutboxMessageParams{
		IdempotencyKey: idempotencyKey,
		Data:           message,
		Kind:           int32(kind),
	})

	if err != nil {
		return fmt.Errorf("send outbox message: %w", err)
	}

	return nil
}

func (o *outboxRepository) GetMessages(
	ctx context.Context,
	batchSize int,
	inProgressTTL time.Duration,
) ([]Data, error) {
	queries := o.getQueries(ctx)
	rows, err := queries.GetOutboxMessages(ctx, sqlc.GetOutboxMessagesParams{
		InProgressTtl: pgtype.Interval{
			Microseconds: inProgressTTL.Microseconds(),
			Valid:        true,
		},
		BatchSize: int32(batchSize),
	})

	if err != nil {
		return nil, fmt.Errorf("get outbox messages: %w", err)
	}

	result := make([]Data, 0, len(rows))
	for _, row := range rows {
		result = append(result, Data{
			IdempotencyKey: row.IdempotencyKey,
			Data:           row.Data,
			Kind:           Kind(row.Kind),
		})
	}

	return result, nil
}

func (o *outboxRepository) MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}
	queries := o.getQueries(ctx)

	err := queries.MarkOutboxMessagesAsProcessed(ctx, idempotencyKeys)
	if err != nil {
		return fmt.Errorf("mark outbox messages as processed: %w", err)
	}

	return nil
}

func (o *outboxRepository) MarkAsRetryable(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}
	queries := o.getQueries(ctx)

	err := queries.MarkOutboxMessagesAsRetryable(ctx, idempotencyKeys)
	if err != nil {
		return fmt.Errorf("mark outbox messages as retryable: %w", err)
	}

	return nil
}
