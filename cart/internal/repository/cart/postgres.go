package cart

import (
	"context"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/entity"
	sqlcCart "github.com/Sushka21/microservices-ecommerce/cart/internal/repository/cart/sqlc"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source=sqlc/querier.go -destination=mocks/querier_mocks.go -package=mocks

type (
	DB interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		sqlcCart.DBTX
	}
)

type postgresRepository struct {
	queries sqlcCart.Querier
	db      DB
}

func NewPostgresRepository(db DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcCart.New(db),
		db:      db,
	}
}

func (r *postgresRepository) getQueries(ctx context.Context) sqlcCart.Querier {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return sqlcCart.New(tx)
	}
	return r.queries
}

func (r *postgresRepository) AddItem(ctx context.Context, userID int64, item entity.CartItem) error {
	queries := r.getQueries(ctx)
	return queries.InsertItem(ctx, sqlcCart.InsertItemParams{
		UserID: userID,
		Sku:    int64(item.SKU),
		Count:  int64(item.Count),
	})
}

func (r *postgresRepository) DeleteItem(ctx context.Context, userID int64, sku uint32) error {
	queries := r.getQueries(ctx)
	return queries.DeleteItemBySku(ctx, sqlcCart.DeleteItemBySkuParams{
		UserID: userID,
		Sku:    int64(sku),
	})
}

func (r *postgresRepository) ListCart(ctx context.Context, userID int64) ([]entity.CartItem, error) {
	queries := r.getQueries(ctx)
	cartDb, err := queries.ListCartByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	cartItem := make([]entity.CartItem, len(cartDb))
	for i, item := range cartDb {
		cartItem[i] = entity.CartItem{
			Count: uint32(item.Count),
			SKU:   uint32(item.Sku),
		}
	}
	return cartItem, nil
}

func (r *postgresRepository) ClearCart(ctx context.Context, userID int64) error {
	return r.getQueries(ctx).ClearCart(ctx, userID)
}
