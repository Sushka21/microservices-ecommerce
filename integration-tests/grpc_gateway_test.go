package tests

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/integration-tests/cart"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/product"
	"github.com/Sushka21/microservices-ecommerce/integration-tests/stocks"
	"github.com/stretchr/testify/require"
)

func TestCartListCart(t *testing.T) {
	cfg, clients := setup(t)

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

	resp, err := jsonReq(http.MethodPost, cfg.Clients.Cart1GatewayAddr+"/v1/cart/list", map[string]int64{"user_id": userID})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest,
		"status %d", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		var body struct {
			Items      []interface{} `json:"items"`
			TotalPrice uint32        `json:"total_price"`
		}
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
	}
}

func TestLOMSOrderInfo(t *testing.T) {
	cfg, _ := setup(t)

	resp, err := jsonReq(http.MethodPost, cfg.Clients.Loms1GatewayAddr+"/v1/order/info", map[string]int64{"order_id": 1})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.True(t, resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusNotFound, "status %d", resp.StatusCode)
}

func TestProductInfo(t *testing.T) {
	cfg, _ := setup(t)

	createResp, err := jsonReq(http.MethodPost, cfg.Clients.Loms1GatewayAddr+"/v1/product/create",
		map[string]interface{}{"name": "Gateway Product", "price": 100})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	var createBody struct {
		Sku uint32 `json:"sku"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createBody)
	require.NoError(t, err)
	require.Greater(t, createBody.Sku, uint32(0))

	infoResp, err := jsonReq(http.MethodPost, cfg.Clients.Loms1GatewayAddr+"/v1/product/info", map[string]uint32{"sku": createBody.Sku})
	require.NoError(t, err)
	defer infoResp.Body.Close()
	require.Equal(t, http.StatusOK, infoResp.StatusCode)
	var infoBody struct {
		Name  string `json:"name"`
		Price uint32 `json:"price"`
	}
	err = json.NewDecoder(infoResp.Body).Decode(&infoBody)
	require.NoError(t, err)
	require.Equal(t, "Gateway Product", infoBody.Name)
	require.Equal(t, uint32(100), infoBody.Price)
}

func TestStocksInfo(t *testing.T) {
	cfg, _ := setup(t)

	createResp, err := jsonReq(http.MethodPost, cfg.Clients.Loms1GatewayAddr+"/v1/product/create",
		map[string]interface{}{"name": "Stock Item", "price": 1})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusOK, createResp.StatusCode)
	var createBody struct {
		Sku uint32 `json:"sku"`
	}
	err = json.NewDecoder(createResp.Body).Decode(&createBody)
	require.NoError(t, err)

	setResp, err := jsonReq(http.MethodPost, cfg.Clients.Loms1GatewayAddr+"/v1/stock/set",
		map[string]interface{}{"sku": createBody.Sku, "count": 42})
	require.NoError(t, err)
	defer setResp.Body.Close()
	require.Equal(t, http.StatusOK, setResp.StatusCode)

	infoResp, err := jsonReq(http.MethodPost, cfg.Clients.Loms1GatewayAddr+"/v1/stock/info",
		map[string]uint32{"sku": createBody.Sku})
	require.NoError(t, err)
	defer infoResp.Body.Close()
	require.Equal(t, http.StatusOK, infoResp.StatusCode)

	var infoBody struct {
		Count uint64OrString `json:"count"`
	}
	err = json.NewDecoder(infoResp.Body).Decode(&infoBody)
	require.NoError(t, err)
	require.Equal(t, uint64(42), uint64(infoBody.Count))
}

type uint64OrString uint64

func (u *uint64OrString) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case float64:
		*u = uint64OrString(uint64(v))
		return nil
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		*u = uint64OrString(n)
		return nil
	default:
		return fmt.Errorf("count: expected number or string, got %T", raw)
	}
}
