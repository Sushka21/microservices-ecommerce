package product

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/product/mocks"
	sqlcproduct "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/product/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPostgresRepository_GetProductBySKU_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	sku := uint32(100)

	querier.EXPECT().
		GetProductBySKU(gomock.Any(), int32(sku)).
		Return(sqlcproduct.LomsProduct{
			Sku:   int32(sku),
			Names: "sneakers",
			Price: 500,
		}, nil)

	productInfo, err := repo.GetProductBySKU(ctx, sku)

	require.NoError(t, err)
	require.Equal(t, entity.ProductInfo{
		Sku:   sku,
		Name:  "sneakers",
		Price: 500,
	}, productInfo)
}

func TestPostgresRepository_GetProductBySKU_Err_Gomock(t *testing.T) {
	t.Parallel()

	getProductErr := errors.New("get product error")

	tests := []struct {
		name        string
		repoErr     error
		expectedErr error
	}{
		{
			name:        "product not found",
			repoErr:     pgx.ErrNoRows,
			expectedErr: entity.ErrProductNotFound,
		},
		{
			name:        "repository error",
			repoErr:     getProductErr,
			expectedErr: getProductErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			querier := mocks.NewMockQuerier(ctrl)

			repo := &postgresRepository{
				queries: querier,
			}

			ctx := context.Background()
			sku := uint32(100)

			querier.EXPECT().
				GetProductBySKU(gomock.Any(), int32(sku)).
				Return(sqlcproduct.LomsProduct{}, tt.repoErr)

			productInfo, err := repo.GetProductBySKU(ctx, sku)

			require.Error(t, err)
			require.Equal(t, entity.ProductInfo{}, productInfo)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_CreateProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()

	productInfo := entity.ProductInfo{
		Name:  "sneakers",
		Price: 500,
	}

	expectedSKU := uint32(100)

	querier.EXPECT().
		CreateProduct(gomock.Any(), sqlcproduct.CreateProductParams{
			Names: productInfo.Name,
			Price: int32(productInfo.Price),
		}).
		Return(sqlcproduct.LomsProduct{
			Sku:   int32(expectedSKU),
			Names: productInfo.Name,
			Price: int32(productInfo.Price),
		}, nil)

	sku, err := repo.CreateProduct(ctx, productInfo)

	require.NoError(t, err)
	require.Equal(t, expectedSKU, sku)
}

func TestPostgresRepository_CreateProduct_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()

	productInfo := entity.ProductInfo{
		Name:  "sneakers",
		Price: 500,
	}

	expectedErr := errors.New("create product error")

	querier.EXPECT().
		CreateProduct(gomock.Any(), sqlcproduct.CreateProductParams{
			Names: productInfo.Name,
			Price: int32(productInfo.Price),
		}).
		Return(sqlcproduct.LomsProduct{}, expectedErr)

	sku, err := repo.CreateProduct(ctx, productInfo)

	require.Error(t, err)
	require.EqualValues(t, 0, sku)
	require.ErrorIs(t, err, expectedErr)
}

func TestPostgresRepository_ListProduct_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	skus := []uint32{100, 200}

	querier.EXPECT().
		ListProductBySkus(gomock.Any(), []int32{100, 200}).
		Return([]sqlcproduct.LomsProduct{
			{
				Sku:   100,
				Names: "sneakers",
				Price: 500,
			},
			{
				Sku:   200,
				Names: "shirt",
				Price: 300,
			},
		}, nil)

	products, err := repo.ListProduct(ctx, skus)

	require.NoError(t, err)
	require.Len(t, products, 2)

	require.Equal(t, entity.ProductInfo{
		Sku:   100,
		Name:  "sneakers",
		Price: 500,
	}, products[0])

	require.Equal(t, entity.ProductInfo{
		Sku:   200,
		Name:  "shirt",
		Price: 300,
	}, products[1])
}

func TestPostgresRepository_ListProduct_Empty_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	products, err := repo.ListProduct(context.Background(), []uint32{})

	require.NoError(t, err)
	require.Empty(t, products)
}

func TestPostgresRepository_ListProduct_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	skus := []uint32{100, 200}
	expectedErr := errors.New("list product error")

	querier.EXPECT().
		ListProductBySkus(gomock.Any(), []int32{100, 200}).
		Return(nil, expectedErr)

	products, err := repo.ListProduct(ctx, skus)

	require.Error(t, err)
	require.Nil(t, products)
	require.ErrorIs(t, err, expectedErr)
}



