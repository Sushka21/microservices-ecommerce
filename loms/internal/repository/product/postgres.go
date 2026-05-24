package product

import (
	"context"
	"errors"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	sqlcproduct "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/product/sqlc"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
)

//go:generate mockgen -source=sqlc/querier.go -destination=mocks/querier_mocks.go -package=mocks

type (
	DB interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		sqlcproduct.DBTX
	}
)

type postgresRepository struct {
	queries sqlcproduct.Querier
	db      DB
}

func NewPostgresRepository(qdb DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcproduct.New(qdb),
		db:      qdb,
	}
}

func (r *postgresRepository) getQueries(ctx context.Context) sqlcproduct.Querier {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return sqlcproduct.New(tx)
	}

	return r.queries
}

func (r *postgresRepository) GetProductBySKU(ctx context.Context, sku uint32) (entity.ProductInfo, error) {
	queries := r.getQueries(ctx)
	productDb, err := queries.GetProductBySKU(ctx, int32((sku)))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ProductInfo{}, entity.ErrProductNotFound
		}
		return entity.ProductInfo{}, err
	}

	return entity.ProductInfo{
		Sku:   sku,
		Name:  productDb.Names,
		Price: uint32(productDb.Price),
	}, nil
}

func (r *postgresRepository) CreateProduct(ctx context.Context, product entity.ProductInfo) (uint32, error) {
	queries := r.getQueries(ctx)
	productDb, err := queries.CreateProduct(ctx, sqlcproduct.CreateProductParams{
		Price: int32(product.Price),
		Names: product.Name,
	})

	if err != nil {
		return 0, err
	}

	return uint32(productDb.Sku), nil
}

func (r *postgresRepository) ListProduct(ctx context.Context, skus []uint32) ([]entity.ProductInfo, error) {
	if len(skus) == 0 {
		return []entity.ProductInfo{}, nil
	}

	queries := r.getQueries(ctx)
	sqlcSkus := make([]int32, len(skus))
	for i, sku := range skus {
		sqlcSkus[i] = int32(sku)
	}

	productDb, err := queries.ListProductBySkus(ctx, sqlcSkus)
	if err != nil {
		return nil, err
	}

	productInfo := make([]entity.ProductInfo, len(productDb))
	for i, product := range productDb {
		productInfo[i] = entity.ProductInfo{
			Sku:   uint32(product.Sku),
			Name:  product.Names,
			Price: uint32(product.Price),
		}
	}

	return productInfo, nil
}
