package kafka

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/port"
	kafka "github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func New(brokers []string, topic string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}

func (p *Publisher) SendMessage(ctx context.Context, userID, orderID int64, status port.OrderStatus) error {
	body := port.OrderStatusChangedNotification{
		OrderID: orderID,
		UserID:  userID,
		Status:  status,
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	// fmt.Printf(
	// 	"LOMS KAFKA SEND userID=%d orderID=%d status=%s key=%s raw=%s\n",
	// 	userID,
	// 	orderID,
	// 	status,
	// 	strconv.FormatInt(orderID, 10),
	// 	string(raw),
	// )

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(orderID, 10)),
		Value: raw,
	})
}
