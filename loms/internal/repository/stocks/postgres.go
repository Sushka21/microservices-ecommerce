package stocks

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	sqlcstocks "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/stocks/sqlc"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const foreignKeyViolation = "23503"

//go:generate mockgen -source=sqlc/querier.go -destination=mocks/querier_mocks.go -package=mocks
type (
	DB interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		sqlcstocks.DBTX
	}
)

type postgresRepository struct {
	queries sqlcstocks.Querier
	db      DB
}

func NewPostgresRepository(qdb DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcstocks.New(qdb),
		db:      qdb,
	}
}

func (r *postgresRepository) getQueries(ctx context.Context) sqlcstocks.Querier {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return sqlcstocks.New(tx)
	}

	return r.queries
}

func (r *postgresRepository) GetCountBySKU(ctx context.Context, sku uint32) (uint64, error) {
	queries := r.getQueries(ctx)
	count, err := queries.GetAvailableStockBySku(ctx, int32(sku))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, entity.ErrProductNotFound
		}
		return 0, err
	}
	return uint64(count), nil
}

func (r *postgresRepository) SetCountBySKU(ctx context.Context, sku uint32, count uint64) error {
	queries := r.getQueries(ctx)
	err := queries.UpsertAvailableStock(ctx, sqlcstocks.UpsertAvailableStockParams{
		Sku:   int32(sku),
		Count: int32(count),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
			return entity.ErrProductNotFound
		}
		return err
	}
	return nil
}

func (r *postgresRepository) ReserveStocks(ctx context.Context, orderID int64, items []entity.OrderItem) error {
	queries := r.getQueries(ctx)
	for _, item := range items {
		check, err := queries.DecrementAvailableStock(ctx, sqlcstocks.DecrementAvailableStockParams{
			Sku:   int32(item.SKU),
			Count: int32(item.Count),
		})
		if err != nil {
			return err
		}
		if check == 0 {
			return entity.ErrInsufficientStock
		}
		err = queries.UpsertReservedStock(ctx, sqlcstocks.UpsertReservedStockParams{
			Sku:     int32(item.SKU),
			Count:   int32(item.Count),
			OrderID: orderID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresRepository) ReleaseStocks(ctx context.Context, orderID int64, items []entity.OrderItem) error {
	queries := r.getQueries(ctx)
	for _, item := range items {
		_, err := queries.DecrementReservedStock(ctx, sqlcstocks.DecrementReservedStockParams{
			Count:   int32(item.Count),
			Sku:     int32(item.SKU),
			OrderID: orderID,
		})
		if err != nil {
			return err
		}
		err = queries.AddAvailableStock(ctx, sqlcstocks.AddAvailableStockParams{
			Sku:   int32(item.SKU),
			Count: int32(item.Count),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresRepository) RemoveReservedStocks(ctx context.Context, orderID int64, items []entity.OrderItem) error {
	queries := r.getQueries(ctx)
	for _, item := range items {
		_, err := queries.DecrementReservedStock(ctx, sqlcstocks.DecrementReservedStockParams{
			Count:   int32(item.Count),
			Sku:     int32(item.SKU),
			OrderID: orderID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
