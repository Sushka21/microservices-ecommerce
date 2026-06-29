package tests

import (
	"math/rand/v2"
	"strconv"
	"sync"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/integration-tests/cart"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/loms"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/stocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestCartStateSharedBetweenInstances(t *testing.T) {
	_, clients := setup(t)

	sku := createProductWithStock(
		t,
		clients.Product1,
		clients.Stocks1,
		"Shared Cart Product",
		100,
		10,
	)

	userID := rand.N[int64](10e9) + 1

	_, err := clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
		UserId: userID,
		Sku:    sku,
		Count:  2,
	})
	require.NoError(t, err)
	requireCartHasItem(t, clients.Cart2, userID, sku, 2)
}

func TestCheckoutWorksAcrossCartInstances(t *testing.T) {
	_, clients := setup(t)

	sku := createProductWithStock(
		t,
		clients.Product1,
		clients.Stocks1,
		"Cross Cart Checkout Product",
		50,
		10,
	)

	userID := rand.N[int64](10e9) + 1

	_, err := clients.Cart1.AddItem(t.Context(), &cart.AddItemRequest{
		UserId: userID,
		Sku:    sku,
		Count:  3,
	})
	require.NoError(t, err)

	checkoutResp, err := clients.Cart2.CheckoutCart(t.Context(), &cart.CheckoutCartRequest{
		UserId: userID,
	})
	require.NoError(t, err)
	require.Greater(t, checkoutResp.GetOrderId(), int64(0))

	orderResp, err := clients.Loms2.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: checkoutResp.GetOrderId(),
	})
	require.NoError(t, err)
	require.Equal(t, userID, orderResp.GetUserId())
	require.Len(t, orderResp.GetItems(), 1)
	require.Equal(t, sku, orderResp.GetItems()[0].GetSku())
	require.EqualValues(t, 3, orderResp.GetItems()[0].GetCount())
}

func TestLOMSOrderSharedBetweenInstances(t *testing.T) {
	_, clients := setup(t)

	sku := createProductWithStock(
		t,
		clients.Product1,
		clients.Stocks1,
		"Shared LOMS Product",
		200,
		100,
	)

	userID := rand.N[int64](10e9) + 1

	createOrderResp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: userID,
		Items: []*loms.Item{
			{Sku: sku, Count: 4},
		},
	})
	require.NoError(t, err)

	orderID := createOrderResp.GetOrderId()
	require.Greater(t, orderID, int64(0))

	getOrderResp, err := clients.Loms2.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)
	require.Equal(t, userID, getOrderResp.GetUserId())
	require.Equal(t, loms.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT, getOrderResp.GetStatus())
	require.Len(t, getOrderResp.GetItems(), 1)
	require.Equal(t, sku, getOrderResp.GetItems()[0].GetSku())
	require.EqualValues(t, 4, getOrderResp.GetItems()[0].GetCount())
}

func TestLOMSStatusChangeVisibleAcrossInstances(t *testing.T) {
	_, clients := setup(t)

	sku := createProductWithStock(
		t,
		clients.Product1,
		clients.Stocks1,
		"Status Sync Product",
		300,
		100,
	)

	createOrderResp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: 777001,
		Items: []*loms.Item{
			{Sku: sku, Count: 2},
		},
	})
	require.NoError(t, err)

	orderID := createOrderResp.GetOrderId()

	_, err = clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)

	getOrderResp, err := clients.Loms2.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)
	require.Equal(t, loms.OrderStatus_ORDER_STATUS_PAID, getOrderResp.GetStatus())
}

func TestConcurrentPayOrderOnlyOneSucceeds(t *testing.T) {
	for iter := range 10 {
		t.Run(strconv.Itoa(iter), func(t *testing.T) {
			_, clients := setup(t)

			_, orderID, _ := createOrderForRaceTest(t, clients, 10, 2)

			start := make(chan struct{})
			errs := make([]error, 2)

			var wg sync.WaitGroup
			for i := range 2 {
				wg.Go(func() {
					<-start
					_, err := clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{
						OrderId: orderID,
					})
					errs[i] = err
				})
			}

			close(start)
			wg.Wait()

			var okCount, failedPreconditionCount int
			for _, err := range errs {
				switch grpcCode(err) {
				case codes.OK:
					okCount++
				case codes.FailedPrecondition:
					failedPreconditionCount++
				default:
					require.NoError(t, err)
				}
			}

			require.Equal(t, 1, okCount)
			require.Equal(t, 1, failedPreconditionCount)

			getResp, err := clients.Loms2.GetOrder(t.Context(), &loms.GetOrderRequest{
				OrderId: orderID,
			})
			require.NoError(t, err)
			require.Equal(t, loms.OrderStatus_ORDER_STATUS_PAID, getResp.GetStatus())
		})
	}
}

func TestConcurrentCancelOrderOnlyOneSucceeds(t *testing.T) {
	_, clients := setup(t)

	sku, orderID, _ := createOrderForRaceTest(t, clients, 10, 3)

	start := make(chan struct{})
	errs := make([]error, 2)

	var wg sync.WaitGroup
	for i := range 2 {
		wg.Go(func() {
			<-start
			_, err := clients.Loms1.CancelOrder(t.Context(), &loms.CancelOrderRequest{
				OrderId: orderID,
			})
			errs[i] = err
		})
	}

	close(start)
	wg.Wait()

	var okCount, failedPreconditionCount int
	for _, err := range errs {
		switch grpcCode(err) {
		case codes.OK:
			okCount++
		case codes.FailedPrecondition:
			failedPreconditionCount++
		default:
			require.NoError(t, err)
		}
	}

	require.Equal(t, 1, okCount)
	require.Equal(t, 1, failedPreconditionCount)

	getResp, err := clients.Loms2.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)
	require.Equal(t, loms.OrderStatus_ORDER_STATUS_CANCELLED, getResp.GetStatus())

	stockResp, err := clients.Stocks2.GetStock(t.Context(), &stocks.GetStockRequest{
		Sku: sku,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(10), stockResp.GetCount())
}

func TestConcurrentPayAndCancelOrderInvariant(t *testing.T) {
	for iter := range 10 {
		t.Run(strconv.Itoa(iter), func(t *testing.T) {
			_, clients := setup(t)

			sku, orderID, _ := createOrderForRaceTest(t, clients, 10, 4)

			start := make(chan struct{})
			errs := make([]error, 2)

			var wg sync.WaitGroup
			wg.Go(func() {
				<-start
				_, errs[0] = clients.Loms1.PayOrder(t.Context(), &loms.PayOrderRequest{
					OrderId: orderID,
				})
			})
			wg.Go(func() {
				<-start
				_, errs[1] = clients.Loms2.CancelOrder(t.Context(), &loms.CancelOrderRequest{
					OrderId: orderID,
				})
			})

			close(start)
			wg.Wait()

			var okCount, failedPreconditionCount int
			for _, err := range errs {
				switch grpcCode(err) {
				case codes.OK:
					okCount++
				case codes.FailedPrecondition:
					failedPreconditionCount++
				default:
					require.NoError(t, err)
				}
			}

			require.Equal(t, 1, okCount)
			require.Equal(t, 1, failedPreconditionCount)

			orderResp, err := clients.Loms1.GetOrder(t.Context(), &loms.GetOrderRequest{
				OrderId: orderID,
			})
			require.NoError(t, err)

			stockResp, err := clients.Stocks1.GetStock(t.Context(), &stocks.GetStockRequest{
				Sku: sku,
			})
			require.NoError(t, err)

			switch orderResp.GetStatus() {
			case loms.OrderStatus_ORDER_STATUS_PAID:
				require.Equal(t, uint64(6), stockResp.GetCount())
			case loms.OrderStatus_ORDER_STATUS_CANCELLED:
				require.Equal(t, uint64(10), stockResp.GetCount())
			default:
				t.Fatalf("unexpected final status: %v", orderResp.GetStatus())
			}
		})
	}
}

func TestConcurrentCreateOrderWithSingleItemInStock(t *testing.T) {
	for iter := range 10 {
		t.Run(strconv.Itoa(iter), func(t *testing.T) {
			_, clients := setup(t)

			sku := createProductWithStock(
				t,
				clients.Product1,
				clients.Stocks1,
				"Single Stock Product",
				100,
				1,
			)

			start := make(chan struct{})
			type result struct {
				orderID int64
				err     error
			}
			results := make([]result, 2)

			var wg sync.WaitGroup
			for i := range 2 {
				wg.Go(func() {
					<-start
					resp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
						UserId: rand.N[int64](10e9) + 1,
						Items: []*loms.Item{
							{Sku: sku, Count: 1},
						},
					})
					if resp != nil {
						results[i].orderID = resp.GetOrderId()
					}
					results[i].err = err
				})
			}

			close(start)
			wg.Wait()

			var okCount, failedPreconditionCount int
			var successfulOrderID int64

			for _, r := range results {
				switch grpcCode(r.err) {
				case codes.OK:
					okCount++
					successfulOrderID = r.orderID
				case codes.FailedPrecondition:
					failedPreconditionCount++
				default:
					require.NoError(t, r.err)
				}
			}

			require.Equal(t, 1, okCount)
			require.Equal(t, 1, failedPreconditionCount)
			require.Greater(t, successfulOrderID, int64(0))

			stockResp, err := clients.Stocks2.GetStock(t.Context(), &stocks.GetStockRequest{
				Sku: sku,
			})
			require.NoError(t, err)
			require.Equal(t, uint64(0), stockResp.GetCount())

			orderResp, err := clients.Loms2.GetOrder(t.Context(), &loms.GetOrderRequest{
				OrderId: successfulOrderID,
			})
			require.NoError(t, err)
			require.Len(t, orderResp.GetItems(), 1)
			require.Equal(t, sku, orderResp.GetItems()[0].GetSku())
			require.EqualValues(t, 1, orderResp.GetItems()[0].GetCount())
		})
	}
}
