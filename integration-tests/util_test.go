package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sushka21/microservices-ecommerce/integration-tests/cart"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/loms"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/product"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/stocks"
	"github.com/caarlos0/env/v10"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type config struct {
	db *sql.DB

	Clients struct {
		Cart1GrpcAddr    string `env:"CART1_GRPC_ADDR" envDefault:"localhost:50051"`
		Cart2GrpcAddr    string `env:"CART2_GRPC_ADDR" envDefault:"localhost:50061"`
		Cart1GatewayAddr string `env:"CART1_GATEWAY_ADDR" envDefault:"localhost:8080"`
		Cart2GatewayAddr string `env:"CART2_GATEWAY_ADDR" envDefault:"localhost:8090"`

		Loms1GrpcAddr    string `env:"LOMS1_GRPC_ADDR" envDefault:"localhost:50052"`
		Loms2GrpcAddr    string `env:"LOMS2_GRPC_ADDR" envDefault:"localhost:50062"`
		Loms1GatewayAddr string `env:"LOMS1_GATEWAY_ADDR" envDefault:"localhost:8081"`
		Loms2GatewayAddr string `env:"LOMS2_GATEWAY_ADDR" envDefault:"localhost:8091"`
	}

	PG struct {
		Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
		Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
		DB       string `env:"POSTGRES_DB" envDefault:"ecommerce"`
		User     string `env:"POSTGRES_USER" envDefault:"ecommerce_user"`
		Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
	}
}

type testClients struct {
	Cart1    cart.CartClient
	Cart2    cart.CartClient
	Loms1    loms.LomsClient
	Loms2    loms.LomsClient
	Product1 product.ProductServiceClient
	Product2 product.ProductServiceClient
	Stocks1  stocks.StocksClient
	Stocks2  stocks.StocksClient
}

func setup(t *testing.T) (*config, *testClients) {
	t.Helper()

	cfg := loadConfig(t)
	cfg.cleanupDB(t)
	waitForServices(t, cfg, 45*time.Second)
	connCart1 := dial(t, cfg.Clients.Cart1GrpcAddr)
	connCart2 := dial(t, cfg.Clients.Cart2GrpcAddr)
	connLoms1 := dial(t, cfg.Clients.Loms1GrpcAddr)
	connLoms2 := dial(t, cfg.Clients.Loms2GrpcAddr)

	return cfg, &testClients{
		Cart1:    cart.NewCartClient(connCart1),
		Cart2:    cart.NewCartClient(connCart2),
		Loms1:    loms.NewLomsClient(connLoms1),
		Loms2:    loms.NewLomsClient(connLoms2),
		Product1: product.NewProductServiceClient(connLoms1),
		Product2: product.NewProductServiceClient(connLoms2),
		Stocks1:  stocks.NewStocksClient(connLoms1),
		Stocks2:  stocks.NewStocksClient(connLoms2),
	}
}

func loadConfig(t *testing.T) *config {
	t.Helper()

	var cfg config
	err := env.Parse(&cfg)
	require.NoError(t, err)

	cfg.Clients.Cart1GatewayAddr = normalizeURL(t, cfg.Clients.Cart1GatewayAddr)
	cfg.Clients.Cart2GatewayAddr = normalizeURL(t, cfg.Clients.Cart2GatewayAddr)

	cfg.Clients.Loms1GatewayAddr = normalizeURL(t, cfg.Clients.Loms1GatewayAddr)
	cfg.Clients.Loms2GatewayAddr = normalizeURL(t, cfg.Clients.Loms2GatewayAddr)

	cfg.initDB(t)

	return &cfg
}

func (c *config) initDB(t *testing.T) {
	t.Helper()

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(c.PG.User),
		url.QueryEscape(c.PG.Password),
		c.PG.Host,
		c.PG.Port,
		c.PG.DB,
	))

	require.NoError(t, err)
	c.db = db
}

func (c *config) cleanupDB(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		require.NoError(t, cleanDB(context.Background(), c.db))
	})
}

func normalizeURL(t *testing.T, u string) string {
	t.Helper()
	return "http://" + strings.TrimLeft(u, "https://")
}

// WaitForCartGateway dirty hack with grpc gateway
func WaitForCartGateway(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := &http.Client{Timeout: 2 * time.Second}
	body := []byte(`{"user_id":1}`)
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Cart gateway %s did not become ready in %v", baseURL, timeout)
		default:
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/cart/list", bytes.NewReader(body))

		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		_ = resp.Body.Close()

		if resp.StatusCode/100 == 2 || resp.StatusCode/100 == 4 {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func WaitForLomsGateway(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := &http.Client{Timeout: 2 * time.Second}
	body := []byte(`{"order_id":1}`)

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("LOMS gateway %s did not become ready in %v", baseURL, timeout)
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/order/info", bytes.NewReader(body))

		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)

		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		_ = resp.Body.Close()

		if resp.StatusCode/100 == 2 || resp.StatusCode/100 == 4 {
			return
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func waitForServices(t *testing.T, cfg *config, timeout time.Duration) {
	t.Helper()
	WaitForCartGateway(t, cfg.Clients.Cart1GatewayAddr, timeout)
	WaitForCartGateway(t, cfg.Clients.Cart2GatewayAddr, timeout)

	WaitForLomsGateway(t, cfg.Clients.Loms1GatewayAddr, timeout) // depends on notifications
	WaitForLomsGateway(t, cfg.Clients.Loms2GatewayAddr, timeout) // depends on notifications
}

func jsonReq(method, url string, body any) (*http.Response, error) {
	var buf bytes.Buffer

	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, &buf)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func dial(t *testing.T, addr string) *grpc.ClientConn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	require.NoError(t, err)

	t.Cleanup(func() {})
	return conn
}

func listCart(t *testing.T, client cart.CartClient, userID int64) []*cart.ListCartResponse {
	t.Helper()
	ctx := context.Background()

	result := make([]*cart.ListCartResponse, 0)
	stream, err := client.ListCart(ctx, &cart.ListCartRequest{UserId: userID})
	require.NoError(t, err)

	for {
		resp, err := stream.Recv()

		if err == io.EOF {
			return result
		}

		require.NoError(t, err)
		result = append(result, resp)
	}
}

func withLock(mutex sync.Locker, action func()) {
	if action == nil {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	action()
}

func cleanDB(ctx context.Context, db *sql.DB) error {
	var tables string

	err := db.QueryRowContext(ctx, `
		SELECT string_agg(
			quote_ident(schemaname) || '.' || quote_ident(tablename),
			', '
		)
		FROM pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		AND tablename NOT IN ('goose_db_version')
	`).Scan(&tables)

	if err != nil {
		return err
	}

	if tables == "" {
		return nil
	}

	query := "TRUNCATE TABLE " + tables + " RESTART IDENTITY CASCADE"

	_, err = db.ExecContext(ctx, query)
	return err
}

func createProductWithStock(
	t *testing.T,
	productClient product.ProductServiceClient,
	stocksClient stocks.StocksClient,
	name string,
	price uint32,
	count uint64,
) uint32 {
	t.Helper()

	createResp, err := productClient.CreateProduct(t.Context(), &product.CreateProductRequest{
		Name:  name,
		Price: price,
	})
	require.NoError(t, err)

	sku := createResp.GetSku()

	_, err = stocksClient.SetStock(t.Context(), &stocks.SetStockRequest{
		Sku:   sku,
		Count: count,
	})
	require.NoError(t, err)

	return sku
}

func requireCartHasItem(
	t *testing.T,
	client cart.CartClient,
	userID int64,
	sku uint32,
	count uint32,
) {
	t.Helper()

	responses := listCart(t, client, userID)
	require.Len(t, responses, 1)
	require.Len(t, responses[0].GetItems(), 1)

	item := responses[0].GetItems()[0]
	require.Equal(t, sku, item.GetSku())
	require.EqualValues(t, count, item.GetCount())
}

func createOrderForRaceTest(
	t *testing.T,
	clients *testClients,
	stock uint64,
	orderCount uint32,
) (sku uint32, orderID int64, userID int64) {
	t.Helper()

	sku = createProductWithStock(
		t,
		clients.Product1,
		clients.Stocks1,
		"Race Product",
		100,
		stock,
	)

	userID = rand.N[int64](10e9) + 1

	resp, err := clients.Loms1.CreateOrder(t.Context(), &loms.CreateOrderRequest{
		UserId: userID,
		Items: []*loms.Item{
			{Sku: sku, Count: orderCount},
		},
	})
	require.NoError(t, err)

	return sku, resp.GetOrderId(), userID
}

func grpcCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	st, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}

	return st.Code()
}
