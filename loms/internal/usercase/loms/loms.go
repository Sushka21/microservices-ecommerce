package loms

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/adapter/notifications/converter"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/port"
	repository "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox"
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mocks.go -package=mocks
type (
	orderRepository interface {
		CreateOrder(ctx context.Context, order entity.Order) (int64, error)
		GetOrderFOrUpdateByID(ctx context.Context, orderID int64) (entity.Order, error)
		GetOrderByID(ctx context.Context, orderID int64) (entity.Order, error)
		SetStatusByID(ctx context.Context, id int64, status entity.OrderStatus) error
	}

	stocksRepository interface {
		ReserveStocks(ctx context.Context, orderID int64, items []entity.OrderItem) error
		ReleaseStocks(ctx context.Context, orderID int64, items []entity.OrderItem) error
	}

	transactor interface {
		WithTx(ctx context.Context, f func(ctx context.Context) error) (err error)
	}
	outboxRepository interface {
		SendMessage(ctx context.Context, idempotencyKey string, kind repository.Kind, message []byte) error
	}
	notificationsClient interface {
		SendMessage(ctx context.Context, userID, orderID int64, status port.OrderStatus) error
	}
)

func NewLomsService(
	orderRepository orderRepository,
	stocksRepository stocksRepository,
	transactor transactor,
	notificationsClient notificationsClient,
	outboxRepository outboxRepository,
) *lomsService {
	return &lomsService{
		orderRepository:     orderRepository,
		stocksRepository:    stocksRepository,
		transactor:          transactor,
		notificationsClient: notificationsClient,
		outboxRepository:    outboxRepository,
	}
}

type lomsService struct {
	orderRepository     orderRepository
	stocksRepository    stocksRepository
	transactor          transactor
	notificationsClient notificationsClient
	outboxRepository    outboxRepository
}

func (s *lomsService) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	var resOrderID int64

	err := s.transactor.WithTx(ctx, func(ctx context.Context) error {
		order := entity.Order{
			UserID:    userID,
			Items:     items,
			Status:    entity.OrderStatusAwaitingPayment,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		orderID, err := s.orderRepository.CreateOrder(ctx, order)
		if err != nil {
			return fmt.Errorf("create order user_id=%d items=%v: %w", userID, items, err)
		}

		if err := s.stocksRepository.ReserveStocks(ctx, orderID, items); err != nil {
			return fmt.Errorf("reserve stocks order_id=%d user_id=%d items=%v: %w", orderID, userID, items, err)
		}

		order.ID = orderID
		resOrderID = orderID

		if err := s.createOutboxStatusChanged(ctx, &order); err != nil {
			return fmt.Errorf("create outbox status changed order_id=%d status=%s: %w", orderID, order.Status, err)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("create order transaction user_id=%d: %w", userID, err)
	}

	return resOrderID, nil
}

func (s *lomsService) GetOrder(ctx context.Context, orderID int64) (entity.Order, error) {
	order, err := s.orderRepository.GetOrderByID(ctx, orderID)
	if err != nil {
		return entity.Order{}, fmt.Errorf("get order by id=%d: %w", orderID, err)
	}
	return order, nil
}

func (s *lomsService) PayOrder(ctx context.Context, orderID int64) error {
	err := s.transactor.WithTx(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.GetOrderFOrUpdateByID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order for update order_id=%d: %w", orderID, err)
		}

		switch order.Status {
		case entity.OrderStatusCancelled:
			return entity.ErrOrderCancelled
		case entity.OrderStatusPaid:
			return entity.ErrOrderAlreadyPaid
		case entity.OrderStatusFailed:
			return entity.ErrOrderFailed

		case entity.OrderStatusNew, entity.OrderStatusAwaitingPayment:
			order.Status = entity.OrderStatusPaid
			order.UpdatedAt = time.Now()

			if err := s.orderRepository.SetStatusByID(ctx, orderID, entity.OrderStatusPaid); err != nil {
				return fmt.Errorf("set paid status order_id=%d: %w", orderID, err)
			}

			if err := s.createOutboxStatusChanged(ctx, &order); err != nil {
				return fmt.Errorf("create outbox status changed order_id=%d status=%s: %w",
					orderID,
					order.Status,
					err,
				)
			}
			return nil

		case entity.OrderStatusUnknown:
			return entity.ErrUnknownOrderStatus
		default:
			return fmt.Errorf("%w: order_id=%d status=%s",
				entity.ErrUnknownOrderStatus,
				orderID,
				order.Status,
			)
		}
	})

	if err != nil {
		return fmt.Errorf("pay order transaction order_id=%d: %w", orderID, err)
	}

	return nil
}

func (s *lomsService) CancelOrder(ctx context.Context, orderID int64) error {
	err := s.transactor.WithTx(ctx, func(ctx context.Context) error {
		order, err := s.orderRepository.GetOrderFOrUpdateByID(ctx, orderID)
		if err != nil {
			return fmt.Errorf("get order for update order_id=%d: %w", orderID, err)
		}

		switch order.Status {
		case entity.OrderStatusCancelled:
			return entity.ErrOrderCancelled
		case entity.OrderStatusPaid:
			return entity.ErrOrderAlreadyPaid
		}

		if err := s.stocksRepository.ReleaseStocks(ctx, orderID, order.Items); err != nil {
			return fmt.Errorf("release stocks order_id=%d items=%v: %w", orderID, order.Items, err)
		}

		order.Status = entity.OrderStatusCancelled
		order.UpdatedAt = time.Now()

		if err := s.orderRepository.SetStatusByID(ctx, orderID, entity.OrderStatusCancelled); err != nil {
			return fmt.Errorf("set cancelled status order_id=%d: %w", orderID, err)
		}

		if err := s.createOutboxStatusChanged(ctx, &order); err != nil {
			return fmt.Errorf("create outbox status changed order_id=%d status=%s: %w",
				orderID,
				order.Status,
				err,
			)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cancel order transaction order_id=%d: %w", orderID, err)
	}

	return nil
}

func (s *lomsService) OrderStatusChangedNotificationKindHandler(ctx context.Context, data []byte) error {
	var body port.OrderStatusChangedNotification
	if err := json.Unmarshal(data, &body); err != nil {
		return fmt.Errorf("unmarshal order status changed notification: %w", err)
	}
	return s.notificationsClient.SendMessage(
		ctx,
		body.UserID,
		body.OrderID,
		body.Status,
	)
}

func (s *lomsService) createOutboxStatusChanged(ctx context.Context, order *entity.Order) error {
	key := s.createKey(order.ID, order.Status)
	body := port.OrderStatusChangedNotification{
		OrderID: order.ID,
		UserID:  order.UserID,
		Status:  converter.FromOrderStatus(order.Status),
	}
	DBbody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal order status changed notification: %w", err)
	}
	return s.outboxRepository.SendMessage(ctx, key, repository.KindNotification, DBbody)
}

func (s *lomsService) createKey(orderID int64, status entity.OrderStatus) string {
	return strconv.FormatInt(orderID, 10) + "_" + string(status)
}



