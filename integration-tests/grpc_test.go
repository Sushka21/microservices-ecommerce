package tests

import (
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/integration-tests/stocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sushka21/microservices-ecommerce/integration-tests/cart"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/loms"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/product"
)

func TestCartOperations(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name:  "Test Product",
		Price: 100,
	})
	require.NoError(t, err)

	sku := createResp.GetSku()
	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{
		Sku:   sku,
		Count: 10,
	},
	)
	require.NoError(t, err)

	var userID = rand.N[int64](10e9) + 1

	_, err = clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
		UserId: userID,
		Sku:    sku, Count: 2,
	},
	)
	require.NoError(t, err)

	cartListResponse := listCart(t, clients.Cart1, userID)
	require.Len(t, cartListResponse, 1)

	cartItems := cartListResponse[0]
	require.Len(t, cartItems.GetItems(), 1)

	require.Equal(t, sku, cartItems.GetItems()[0].GetSku())
	require.EqualValues(t, 2, cartItems.GetItems()[0].GetCount())

	_, err = clients.Cart1.DeleteItem(t.Context(), &cart.DeleteItemRequest{
		UserId: userID,
		Sku:    sku},
	)
	require.NoError(t, err)

	cartListResponseAfterDelete := listCart(t, clients.Cart1, userID)
	require.Zero(t, len(cartListResponseAfterDelete))

	_, err = clients.Cart1.ClearCart(t.Context(), &cart.ClearCartRequest{
		UserId: userID,
	})
	require.NoError(t, err)
}

func TestCartAddItemProductNotFound(t *testing.T) {
	_, clients := setup(t)
	_, err := clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
		UserId: rand.N[int64](10e9),
		Sku:    999999999,
		Count:  1},
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Contains(t, []codes.Code{codes.NotFound, codes.InvalidArgument, codes.FailedPrecondition}, st.Code())
}

func TestCartAddItemInsufficientStock(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name:  "Low Stock",
		Price: 1},
	)
	require.NoError(t, err)
	sku := createResp.GetSku()
	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{Sku: sku, Count: 1})
	require.NoError(t, err)

	_, err = clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
		UserId: rand.N[int64](10e9),
		Sku:    sku,
		Count:  10,
	},
	)

	require.Error(t, err)
	st, ok := status.FromError(err)

	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestConcurrentStocksAndCartConsistency(t *testing.T) {
	_, clients := setup(t)

	const numProducts = 5
	const stockPerProduct = 500
	const numWorkers = 12
	const addOpsPerWorker = 40

	skus := make([]uint32, numProducts)
	for i := 0; i < numProducts; i++ {
		createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
			Name:  "Concurrent Product " + string(rune('A'+i)),
			Price: uint32(10 + i),
		})
		require.NoError(t, err)
		skus[i] = createResp.GetSku()
	}

	var setStockWg sync.WaitGroup
	for i := 0; i < numProducts; i++ {
		setStockWg.Go(func() {
			_, err := clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{
				Sku:   skus[i],
				Count: stockPerProduct,
			})
			require.NoError(t, err)
		})
	}

	setStockWg.Wait()

	type countMap map[uint32]uint32
	expected := make(map[int64]countMap)
	expectedMx := new(sync.Mutex)
	cartWg := new(sync.WaitGroup)

	for w := 0; w < numWorkers; w++ {
		userID := rand.N[int64](10e9) + int64(w+1)*1e8
		cartWg.Go(func() {
			localCounts := make(countMap)
			for op := 0; op < addOpsPerWorker; op++ {
				sku := skus[rand.N(len(skus))]
				count := uint32(rand.N(2) + 1)
				_, err := clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
					UserId: userID,
					Sku:    sku,
					Count:  count,
				})
				if err != nil {
					continue
				}
				localCounts[sku] += count
			}

			withLock(expectedMx, func() {
				expected[userID] = localCounts

			})
		})

	}
	cartWg.Wait()

	for userID, expectedCounts := range expected {
		responses := listCart(t, clients.Cart1, userID)
		gotCounts := make(countMap)
		for _, r := range responses {
			for _, it := range r.GetItems() {
				gotCounts[it.GetSku()] += it.GetCount()
			}
		}

		for sku, want := range expectedCounts {
			require.Equal(t, want, gotCounts[sku], "user %d sku %d", userID, sku)
		}

		for sku, got := range gotCounts {
			require.Equal(t, expectedCounts[sku], got, "user %d sku %d", userID, sku)
		}
	}

	for _, sku := range skus {
		resp, err := clients.Stocks1.GetStock(t.Context(), &stocks.GetStockRequest{Sku: sku})
		require.NoError(t, err)
		require.Equal(t, uint64(stockPerProduct), resp.GetCount(), "sku %d stock", sku)
	}
}

func TestCheckoutCart(t *testing.T) {
	_, clients := setup(t)
	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name:  "Checkout Product",
		Price: 50,
	})
	require.NoError(t, err)

	sku := createResp.GetSku()
	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{Sku: sku, Count: 5})
	require.NoError(t, err)

	userID := rand.N[int64](10e9)
	_, err = clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
		UserId: userID,
		Sku:    sku,
		Count:  2,
	})
	require.NoError(t, err)

	checkoutResp, err := clients.Cart1.CheckoutCart(t.Context(), &cart.CheckoutCartRequest{
		UserId: userID,
	})
	require.NoError(t, err)
	require.Greater(t, checkoutResp.GetOrderId(), int64(0))
}

func TestLOMSOrderOperations(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name:  "Order Product",
		Price: 200,
	})
	require.NoError(t, err)
	sku := createResp.GetSku()

	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{Sku: sku, Count: 100})
	require.NoError(t, err)

	var userID = rand.N[int64](10e9)

	orderResp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: userID,
		Items:  []*loms.Item{{Sku: sku, Count: 3}},
	})
	require.NoError(t, err)
	orderID := orderResp.GetOrderId()
	require.Greater(t, orderID, int64(0))

	getResp, err := clients.Loms1.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)
	require.Equal(t, loms.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT, getResp.GetStatus())
	require.Equal(t, userID, getResp.GetUserId())
	require.Len(t, getResp.GetItems(), 1)

	_, err = clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)

	getResp2, err := clients.Loms1.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)
	require.Equal(t, loms.OrderStatus_ORDER_STATUS_PAID, getResp2.GetStatus())
}

func TestGRPC_LOMS_GetOrder_NotFound(t *testing.T) {
	_, clients := setup(t)
	_, err := clients.Loms1.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: 27272727272727,
	})

	require.Error(t, err)
	st, ok := status.FromError(err)

	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestLOMSPayOrderInvalidStatus(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name: "Pay Twice", Price: 1,
	})
	require.NoError(t, err)

	sku := createResp.GetSku()
	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{
		Sku:   sku,
		Count: 10,
	})
	require.NoError(t, err)

	orderResp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: 999030,
		Items:  []*loms.Item{{Sku: sku, Count: 1}}})
	require.NoError(t, err)
	orderID := orderResp.GetOrderId()

	_, err = clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{OrderId: orderID})
	require.NoError(t, err)

	_, err = clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{OrderId: orderID})
	require.Error(t, err)
	st, ok := status.FromError(err)

	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestLOMSCancelOrderInvalidStatus(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name: "Cancel Paid", Price: 1,
	})
	require.NoError(t, err)

	sku := createResp.GetSku()
	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{
		Sku: sku, Count: 10,
	})
	require.NoError(t, err)

	orderResp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: 999040, Items: []*loms.Item{{Sku: sku, Count: 1}},
	})
	require.NoError(t, err)

	orderID := orderResp.GetOrderId()
	_, err = clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{OrderId: orderID})

	require.NoError(t, err)

	_, err = clients.Loms1.CancelOrder(t.Context(), &loms.CancelOrderRequest{OrderId: orderID})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestLOMSCreateOrderInsufficientStock(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{Name: "No Stock", Price: 1})
	require.NoError(t, err)
	sku := createResp.GetSku()

	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{Sku: sku, Count: 0})
	require.NoError(t, err)

	_, err = clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: 999050,
		Items:  []*loms.Item{{Sku: sku, Count: 5}},
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
}

func TestLOMSCancelOrderReleasesStock(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name:  "Cancel Stock",
		Price: 1,
	})
	require.NoError(t, err)

	sku := createResp.GetSku()
	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{
		Sku:   sku,
		Count: 10,
	})
	require.NoError(t, err)

	orderResp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: 999060,
		Items:  []*loms.Item{{Sku: sku, Count: 3}},
	})
	require.NoError(t, err)
	orderID := orderResp.GetOrderId()
	stockAfterReserve, _ := clients.Stocks1.GetStock(t.Context(), &stocks.GetStockRequest{Sku: sku})
	require.Equal(t, uint64(7), stockAfterReserve.GetCount())

	_, err = clients.Loms1.CancelOrder(t.Context(), &loms.CancelOrderRequest{OrderId: orderID})
	require.NoError(t, err)
	stockAfterCancel, _ := clients.Stocks1.GetStock(t.Context(), &stocks.GetStockRequest{Sku: sku})
	require.Equal(t, uint64(10), stockAfterCancel.GetCount())
}

func TestProductOperations(t *testing.T) {
	_, clients := setup(t)

	createResp, err := clients.Product1.CreateProduct(t.Context(), &product.CreateProductRequest{Name: "Proto Test", Price: 42})
	require.NoError(t, err)
	sku := createResp.GetSku()
	getResp, err := clients.Product1.GetProduct(t.Context(), &product.GetProductRequest{Sku: sku})
	require.NoError(t, err)
	require.Equal(t, "Proto Test", getResp.GetName())
	require.Equal(t, uint32(42), getResp.GetPrice())

	_, err = clients.Stocks1.SetStock(t.Context(), &stocks.SetStockRequest{Sku: sku, Count: 77})
	require.NoError(t, err)
	getStockResp, err := clients.Stocks1.GetStock(t.Context(), &stocks.GetStockRequest{Sku: sku})
	require.NoError(t, err)
	require.Equal(t, uint64(77), getStockResp.GetCount())
}



