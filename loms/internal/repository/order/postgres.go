package order

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	entity "github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	sqlcorder "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/order/sqlc"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/transactor"
)

//go:generate mockgen -source=sqlc/querier.go -destination=mocks/querier_mocks.go -package=mocks

type (
	DB interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		sqlcorder.DBTX
	}
)

type postgresRepository struct {
	queries sqlcorder.Querier
	db      DB
}

func NewPostgresRepository(qdb DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcorder.New(qdb),
		db:      qdb,
	}
}

func (r *postgresRepository) getQueries(ctx context.Context) sqlcorder.Querier {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return sqlcorder.New(tx)
	}

	return r.queries
}

func (r *postgresRepository) CreateOrder(
	ctx context.Context,
	o entity.Order,
) (int64, error) {
	queries := r.getQueries(ctx)
	order, err := queries.InsertOrder(ctx, sqlcorder.InsertOrderParams{
		UserID: o.UserID,
		Status: toSQLCStatus(o.Status),
	})
	if err != nil {
		return 0, err
	}
	for _, item := range o.Items {
		err = queries.InsertOrderItem(ctx, sqlcorder.InsertOrderItemParams{
			OrderID: order.ID,
			Sku:     int64(item.SKU),
			Count:   int64(item.Count),
		})
		if err != nil {
			return 0, err
		}
	}
	return order.ID, nil
}

func (r *postgresRepository) GetOrderByID(ctx context.Context, orderID int64) (entity.Order, error) {
	queries := r.getQueries(ctx)
	order, err := queries.GetOrderByID(ctx, orderID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Order{}, entity.ErrOrderNotFound
		}

		return entity.Order{}, err
	}

	itemsBd, err := queries.ListOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return entity.Order{}, err
	}
	items := make([]entity.OrderItem, 0, len(itemsBd))
	for _, item := range itemsBd {
		items = append(items, entity.OrderItem{
			SKU:   uint32(item.Sku),
			Count: uint64(item.Count),
		})
	}
	return entity.Order{
		ID:        order.ID,
		UserID:    order.UserID,
		Items:     items,
		Status:    toEntityStatus(order.Status),
		CreatedAt: order.CreatedAt.Time,
		UpdatedAt: order.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) GetOrderFOrUpdateByID(ctx context.Context, orderID int64) (entity.Order, error) {
	queries := r.getQueries(ctx)
	order, err := queries.GetOrderForUpdateByID(ctx, orderID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Order{}, entity.ErrOrderNotFound
		}

		return entity.Order{}, err
	}

	itemsBd, err := queries.ListOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return entity.Order{}, err
	}
	items := make([]entity.OrderItem, 0, len(itemsBd))
	for _, item := range itemsBd {
		items = append(items, entity.OrderItem{
			SKU:   uint32(item.Sku),
			Count: uint64(item.Count),
		})
	}
	return entity.Order{
		ID:        order.ID,
		UserID:    order.UserID,
		Items:     items,
		Status:    toEntityStatus(order.Status),
		CreatedAt: order.CreatedAt.Time,
		UpdatedAt: order.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) SetStatusByID(
	ctx context.Context,
	orderID int64,
	status entity.OrderStatus,
) error {
	queries := r.getQueries(ctx)
	return queries.SetOrderStatus(ctx, sqlcorder.SetOrderStatusParams{
		Status: toSQLCStatus(status),
		ID:     orderID,
	})
}

func toSQLCStatus(status entity.OrderStatus) sqlcorder.LomsOrderStatus {
	switch status {
	case entity.OrderStatusNew:
		return sqlcorder.LomsOrderStatusNew
	case entity.OrderStatusAwaitingPayment:
		return sqlcorder.LomsOrderStatusAwaitingpayment
	case entity.OrderStatusFailed:
		return sqlcorder.LomsOrderStatusFailed
	case entity.OrderStatusPaid:
		return sqlcorder.LomsOrderStatusPaid
	case entity.OrderStatusCancelled:
		return sqlcorder.LomsOrderStatusCancelled
	default:
		return sqlcorder.LomsOrderStatusNew
	}
}

func toEntityStatus(status sqlcorder.LomsOrderStatus) entity.OrderStatus {
	switch status {
	case sqlcorder.LomsOrderStatusNew:
		return entity.OrderStatusNew
	case sqlcorder.LomsOrderStatusAwaitingpayment:
		return entity.OrderStatusAwaitingPayment
	case sqlcorder.LomsOrderStatusFailed:
		return entity.OrderStatusFailed
	case sqlcorder.LomsOrderStatusPaid:
		return entity.OrderStatusPaid
	case sqlcorder.LomsOrderStatusCancelled:
		return entity.OrderStatusCancelled
	default:
		return entity.OrderStatusNew
	}
}
