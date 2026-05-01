package tests

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Sushka21/microservices-ecommerce/integration-tests/loms"
	"github.com/stretchr/testify/require"
)

type callbackRequest struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

func TestCreateOrderSucceedsWhenNotificationCallbackFails(t *testing.T) {
	ensureCallbackServer(t)
	callbackStore.SetFail(true)

	_, clients := setup(t)

	orderID := createOrderForNotificationTest(t, clients)
	require.Greater(t, orderID, int64(0))

	getResp, err := clients.Loms1.GetOrder(t.Context(), &loms.GetOrderRequest{
		OrderId: orderID,
	})
	require.NoError(t, err)
	require.Equal(t, loms.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT, getResp.GetStatus())

	require.Eventually(t, func() bool {
		return callbackStore.attemptsByOrder(orderID) >= 1
	}, 5*time.Second, 100*time.Millisecond)
}

func TestOutboxRetriesAndEventuallyDeliversNotification(t *testing.T) {
	ensureCallbackServer(t)
	callbackStore.SetFail(true)

	_, clients := setup(t)

	orderID := createOrderForNotificationTest(t, clients)

	require.Eventually(t, func() bool {
		return callbackStore.attemptsByOrder(orderID) >= 1
	}, 5*time.Second, 100*time.Millisecond)

	require.Equal(t, 0, callbackStore.successesByOrder(orderID))

	callbackStore.SetFail(false)

	require.Eventually(t, func() bool {
		return callbackStore.successesByOrder(orderID) == 1
	}, 8*time.Second, 100*time.Millisecond)
}

func TestOutboxDoesNotSendDuplicateNotificationAfterSuccess(t *testing.T) {
	ensureCallbackServer(t)
	callbackStore.SetFail(true)

	_, clients := setup(t)

	orderID := createOrderForNotificationTest(t, clients)

	require.Eventually(t, func() bool {
		return callbackStore.attemptsByOrder(orderID) >= 1
	}, 5*time.Second, 100*time.Millisecond)

	callbackStore.SetFail(false)

	require.Eventually(t, func() bool {
		return callbackStore.successesByOrder(orderID) == 1
	}, 8*time.Second, 100*time.Millisecond)

	time.Sleep(5 * time.Second)

	require.Equal(t, 1, callbackStore.successesByOrder(orderID))
}

type callbackRecorder struct {
	mx sync.Mutex

	fail bool

	attempts  []callbackRequest
	successes []callbackRequest
}

func (r *callbackRecorder) Reset() {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.fail = false
	r.attempts = nil
	r.successes = nil
}

func (r *callbackRecorder) SetFail(v bool) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.fail = v
}

func (r *callbackRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()

	var payload callbackRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.mx.Lock()
	r.attempts = append(r.attempts, payload)
	shouldFail := r.fail
	if !shouldFail {
		r.successes = append(r.successes, payload)
	}
	r.mx.Unlock()

	if shouldFail {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("temporary callback failure"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (r *callbackRecorder) attemptsByOrder(orderID int64) int {
	r.mx.Lock()
	defer r.mx.Unlock()

	count := 0
	for _, item := range r.attempts {
		if item.OrderID == orderID {
			count++
		}
	}

	return count
}

func (r *callbackRecorder) successesByOrder(orderID int64) int {
	r.mx.Lock()
	defer r.mx.Unlock()

	count := 0
	for _, item := range r.successes {
		if item.OrderID == orderID {
			count++
		}
	}

	return count
}

var (
	callbackSrvOnce sync.Once
	callbackStore   = &callbackRecorder{}
)

func ensureCallbackServer(t *testing.T) {
	t.Helper()

	callbackSrvOnce.Do(func() {
		mux := http.NewServeMux()
		mux.Handle("/", callbackStore)

		ln, err := net.Listen("tcp", ":8080")
		require.NoError(t, err)

		srv := &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 2 * time.Second,
		}

		go func() {
			_ = srv.Serve(ln)
		}()
	})

	callbackStore.Reset()
}

func createOrderForNotificationTest(t *testing.T, clients *testClients) (orderID int64) {
	t.Helper()

	sku := createProductWithStock(
		t,
		clients.Product1,
		clients.Stocks1,
		"Notification Product",
		100,
		10,
	)

	resp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: 900001,
		Items: []*loms.Item{
			{Sku: sku, Count: 2},
		},
	})
	require.NoError(t, err)
	require.Greater(t, resp.GetOrderId(), int64(0))

	return resp.GetOrderId()
}



