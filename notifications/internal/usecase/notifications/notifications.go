package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	"go.uber.org/zap"
)

var (
	ErrCallbackRequestFailed = errors.New("callback request failed")
	ErrCallbackBadStatus     = errors.New("callback returned bad status")
)

type notificationsService struct {
	logger *zap.Logger
	cfg    *config.Config
	client *http.Client
}

type callbackPayload struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

func NewNotificationsService(logger *zap.Logger, cfg *config.Config) *notificationsService {
	return &notificationsService{
		logger: logger,
		cfg:    cfg,
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

func (s *notificationsService) SendMessage(ctx context.Context, userID, orderID int64, status string) error {
	callBackAddr := strings.TrimSpace(s.cfg.Clients.CallbackAddr)
	if callBackAddr == "" {
		s.logger.Warn("callback address is empty, skipping notification",
			zap.Int64("user_id", userID),
			zap.Int64("order_id", orderID),
			zap.String("status", status),
		)
		return nil
	}

	if !strings.HasPrefix(callBackAddr, "http://") && !strings.HasPrefix(callBackAddr, "https://") {
		callBackAddr = "http://" + callBackAddr
	}

	payload := &callbackPayload{
		UserID:  userID,
		OrderID: orderID,
		Status:  status,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal callback payload",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.Int64("order_id", orderID),
			zap.String("status", status),
		)
		return fmt.Errorf("marshal callback payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callBackAddr, bytes.NewReader(body))
	if err != nil {
		s.logger.Error("failed to create callback request",
			zap.Error(err),
			zap.String("callback_addr", callBackAddr),
			zap.Int64("user_id", userID),
			zap.Int64("order_id", orderID),
			zap.String("status", status),
		)
		return fmt.Errorf("%w: %v", ErrCallbackRequestFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("failed to send callback notification",
			zap.Error(err),
			zap.String("callback_addr", callBackAddr),
			zap.Int64("user_id", userID),
			zap.Int64("order_id", orderID),
			zap.String("status", status),
		)
		return fmt.Errorf("%w: %v", ErrCallbackRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		s.logger.Error("callback notification failed",
			zap.String("callback_addr", callBackAddr),
			zap.Int("status_code", resp.StatusCode),
			zap.Int64("user_id", userID),
			zap.Int64("order_id", orderID),
			zap.String("status", status),
		)
		return fmt.Errorf("%w: status_code=%d", ErrCallbackBadStatus, resp.StatusCode)
	}
	return nil
}



