package usecase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNotificationsService_SendMessage_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedUserID := int64(1)
	expectedOrderID := int64(10)
	expectedStatus := "paid"

	requestReceived := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload callbackPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		require.Equal(t, expectedUserID, payload.UserID)
		require.Equal(t, expectedOrderID, payload.OrderID)
		require.Equal(t, expectedStatus, payload.Status)

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	srv := NewNotificationsService(zap.NewNop(), &config.Config{})
	srv.cfg.Clients.CallbackAddr = server.URL

	// Act
	err := srv.SendMessage(context.Background(), expectedUserID, expectedOrderID, expectedStatus)

	// Assert
	require.NoError(t, err)
	require.True(t, requestReceived)
}

func TestNotificationsService_SendMessage_EmptyCallbackAddr(t *testing.T) {
	t.Parallel()

	// Arrange
	srv := NewNotificationsService(zap.NewNop(), &config.Config{})
	srv.cfg.Clients.CallbackAddr = ""

	// Act
	err := srv.SendMessage(context.Background(), 1, 10, "paid")

	// Assert
	require.NoError(t, err)
}

func TestNotificationsService_SendMessage_AddsHTTPPrefix(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedUserID := int64(1)
	expectedOrderID := int64(10)
	expectedStatus := "paid"

	requestReceived := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		var payload callbackPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		require.Equal(t, expectedUserID, payload.UserID)
		require.Equal(t, expectedOrderID, payload.OrderID)
		require.Equal(t, expectedStatus, payload.Status)

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	srv := NewNotificationsService(zap.NewNop(), &config.Config{})

	srv.cfg.Clients.CallbackAddr = server.URL[len("http://"):]

	// Act
	err := srv.SendMessage(context.Background(), expectedUserID, expectedOrderID, expectedStatus)

	// Assert
	require.NoError(t, err)
	require.True(t, requestReceived)
}

func TestNotificationsService_SendMessage_BadStatus(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	srv := NewNotificationsService(zap.NewNop(), &config.Config{})
	srv.cfg.Clients.CallbackAddr = server.URL

	// Act
	err := srv.SendMessage(context.Background(), 1, 10, "paid")

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCallbackBadStatus)
	require.Contains(t, err.Error(), "status_code=500")
}

func TestNotificationsService_SendMessage_RequestFailed(t *testing.T) {
	t.Parallel()

	// Arrange
	srv := NewNotificationsService(zap.NewNop(), &config.Config{})
	srv.cfg.Clients.CallbackAddr = "http://127.0.0.1:1"

	// Act
	err := srv.SendMessage(context.Background(), 1, 10, "paid")

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCallbackRequestFailed)
}

func TestNotificationsService_SendMessage_InvalidCallbackURL(t *testing.T) {
	t.Parallel()

	// Arrange
	srv := NewNotificationsService(zap.NewNop(), &config.Config{})

	srv.cfg.Clients.CallbackAddr = "://bad-url"

	// Act
	err := srv.SendMessage(context.Background(), 1, 10, "paid")

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCallbackRequestFailed)
}

func TestNotificationsService_SendMessage_ContextCanceled(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	srv := NewNotificationsService(zap.NewNop(), &config.Config{})
	srv.cfg.Clients.CallbackAddr = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err := srv.SendMessage(ctx, 1, 10, "paid")

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCallbackRequestFailed)
}



