package stocks

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/stocks/mocks"
	sqlcstocks "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/stocks/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPostgresRepository_GetCountBySKU_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	sku := uint32(100)
	expectedCount := int32(10)

	querier.EXPECT().
		GetAvailableStockBySku(gomock.Any(), int32(sku)).
		Return(expectedCount, nil)

	count, err := repo.GetCountBySKU(ctx, sku)

	require.NoError(t, err)
	require.EqualValues(t, expectedCount, count)
}

func TestPostgresRepository_GetCountBySKU_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("get stock error")

	tests := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{
			name:        "product not found",
			err:         pgx.ErrNoRows,
			expectedErr: entity.ErrProductNotFound,
		},
		{
			name:        "repository error",
			err:         repoErr,
			expectedErr: repoErr,
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

			sku := uint32(100)

			querier.EXPECT().
				GetAvailableStockBySku(gomock.Any(), int32(sku)).
				Return(int32(0), tt.err)

			count, err := repo.GetCountBySKU(context.Background(), sku)

			require.Error(t, err)
			require.EqualValues(t, 0, count)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_SetCountBySKU_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	sku := uint32(100)
	count := uint64(10)

	querier.EXPECT().
		UpsertAvailableStock(gomock.Any(), sqlcstocks.UpsertAvailableStockParams{
			Sku:   int32(sku),
			Count: int32(count),
		}).
		Return(nil)

	err := repo.SetCountBySKU(ctx, sku, count)

	require.NoError(t, err)
}

func TestPostgresRepository_SetCountBySKU_Err_Gomock(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("upsert stock error")

	tests := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{
			name: "foreign key product not found",
			err: &pgconn.PgError{
				Code: "23503",
			},
			expectedErr: entity.ErrProductNotFound,
		},
		{
			name:        "repository error",
			err:         repoErr,
			expectedErr: repoErr,
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

			sku := uint32(100)
			count := uint64(10)

			querier.EXPECT().
				UpsertAvailableStock(gomock.Any(), sqlcstocks.UpsertAvailableStockParams{
					Sku:   int32(sku),
					Count: int32(count),
				}).
				Return(tt.err)

			err := repo.SetCountBySKU(context.Background(), sku, count)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_ReserveStocks_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)

	items := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	gomock.InOrder(
		querier.EXPECT().
			DecrementAvailableStock(gomock.Any(), sqlcstocks.DecrementAvailableStockParams{
				Sku:   100,
				Count: 2,
			}).
			Return(int64(1), nil),

		querier.EXPECT().
			UpsertReservedStock(gomock.Any(), sqlcstocks.UpsertReservedStockParams{
				Sku:     100,
				Count:   2,
				OrderID: orderID,
			}).
			Return(nil),

		querier.EXPECT().
			DecrementAvailableStock(gomock.Any(), sqlcstocks.DecrementAvailableStockParams{
				Sku:   200,
				Count: 1,
			}).
			Return(int64(1), nil),

		querier.EXPECT().
			UpsertReservedStock(gomock.Any(), sqlcstocks.UpsertReservedStockParams{
				Sku:     200,
				Count:   1,
				OrderID: orderID,
			}).
			Return(nil),
	)

	err := repo.ReserveStocks(ctx, orderID, items)

	require.NoError(t, err)
}

func TestPostgresRepository_ReserveStocks_EmptyItems_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	err := repo.ReserveStocks(context.Background(), 10, nil)

	require.NoError(t, err)
}

func TestPostgresRepository_ReserveStocks_Err_Gomock(t *testing.T) {
	t.Parallel()

	decrementErr := errors.New("decrement available stock error")
	upsertReservedErr := errors.New("upsert reserved stock error")

	tests := []struct {
		name         string
		check        int64
		decrementErr error
		upsertErr    error
		expectedErr  error
	}{
		{
			name:         "decrement available stock error",
			decrementErr: decrementErr,
			expectedErr:  decrementErr,
		},
		{
			name:        "insufficient stock",
			check:       0,
			expectedErr: entity.ErrInsufficientStock,
		},
		{
			name:        "upsert reserved stock error",
			check:       1,
			upsertErr:   upsertReservedErr,
			expectedErr: upsertReservedErr,
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

			orderID := int64(10)
			items := []entity.OrderItem{
				{
					SKU:   100,
					Count: 2,
				},
			}

			querier.EXPECT().
				DecrementAvailableStock(gomock.Any(), sqlcstocks.DecrementAvailableStockParams{
					Sku:   100,
					Count: 2,
				}).
				Return(tt.check, tt.decrementErr)

			if tt.decrementErr == nil && tt.check != 0 {
				querier.EXPECT().
					UpsertReservedStock(gomock.Any(), sqlcstocks.UpsertReservedStockParams{
						Sku:     100,
						Count:   2,
						OrderID: orderID,
					}).
					Return(tt.upsertErr)
			}

			err := repo.ReserveStocks(context.Background(), orderID, items)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_ReleaseStocks_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)

	items := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	gomock.InOrder(
		querier.EXPECT().
			DecrementReservedStock(gomock.Any(), sqlcstocks.DecrementReservedStockParams{
				Sku:     100,
				Count:   2,
				OrderID: orderID,
			}).
			Return(int64(1), nil),

		querier.EXPECT().
			AddAvailableStock(gomock.Any(), sqlcstocks.AddAvailableStockParams{
				Sku:   100,
				Count: 2,
			}).
			Return(nil),

		querier.EXPECT().
			DecrementReservedStock(gomock.Any(), sqlcstocks.DecrementReservedStockParams{
				Sku:     200,
				Count:   1,
				OrderID: orderID,
			}).
			Return(int64(1), nil),

		querier.EXPECT().
			AddAvailableStock(gomock.Any(), sqlcstocks.AddAvailableStockParams{
				Sku:   200,
				Count: 1,
			}).
			Return(nil),
	)

	err := repo.ReleaseStocks(ctx, orderID, items)

	require.NoError(t, err)
}

func TestPostgresRepository_ReleaseStocks_EmptyItems_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	err := repo.ReleaseStocks(context.Background(), 10, []entity.OrderItem{})

	require.NoError(t, err)
}

func TestPostgresRepository_ReleaseStocks_Err_Gomock(t *testing.T) {
	t.Parallel()

	decrementErr := errors.New("decrement reserved stock error")
	addErr := errors.New("add available stock error")

	tests := []struct {
		name         string
		decrementErr error
		addErr       error
		expectedErr  error
	}{
		{
			name:         "decrement reserved stock error",
			decrementErr: decrementErr,
			expectedErr:  decrementErr,
		},
		{
			name:        "add available stock error",
			addErr:      addErr,
			expectedErr: addErr,
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

			orderID := int64(10)
			items := []entity.OrderItem{
				{
					SKU:   100,
					Count: 2,
				},
			}

			querier.EXPECT().
				DecrementReservedStock(gomock.Any(), sqlcstocks.DecrementReservedStockParams{
					Sku:     100,
					Count:   2,
					OrderID: orderID,
				}).
				Return(int64(1), tt.decrementErr)

			if tt.decrementErr == nil {
				querier.EXPECT().
					AddAvailableStock(gomock.Any(), sqlcstocks.AddAvailableStockParams{
						Sku:   100,
						Count: 2,
					}).
					Return(tt.addErr)
			}

			err := repo.ReleaseStocks(context.Background(), orderID, items)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPostgresRepository_RemoveReservedStocks_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)

	items := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
		{
			SKU:   200,
			Count: 1,
		},
	}

	gomock.InOrder(
		querier.EXPECT().
			DecrementReservedStock(gomock.Any(), sqlcstocks.DecrementReservedStockParams{
				Sku:     100,
				Count:   2,
				OrderID: orderID,
			}).
			Return(int64(1), nil),

		querier.EXPECT().
			DecrementReservedStock(gomock.Any(), sqlcstocks.DecrementReservedStockParams{
				Sku:     200,
				Count:   1,
				OrderID: orderID,
			}).
			Return(int64(1), nil),
	)

	err := repo.RemoveReservedStocks(ctx, orderID, items)

	require.NoError(t, err)
}

func TestPostgresRepository_RemoveReservedStocks_EmptyItems_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	err := repo.RemoveReservedStocks(context.Background(), 10, nil)

	require.NoError(t, err)
}

func TestPostgresRepository_RemoveReservedStocks_Err_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	querier := mocks.NewMockQuerier(ctrl)

	repo := &postgresRepository{
		queries: querier,
	}

	ctx := context.Background()
	orderID := int64(10)

	items := []entity.OrderItem{
		{
			SKU:   100,
			Count: 2,
		},
	}

	expectedErr := errors.New("decrement reserved stock error")

	querier.EXPECT().
		DecrementReservedStock(gomock.Any(), sqlcstocks.DecrementReservedStockParams{
			Sku:     100,
			Count:   2,
			OrderID: orderID,
		}).
		Return(int64(0), expectedErr)

	err := repo.RemoveReservedStocks(ctx, orderID, items)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}



