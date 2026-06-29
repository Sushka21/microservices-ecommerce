package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"go.uber.org/zap"
)

//go:generate mockgen -source=consumer.go -destination=mocks/consumer_mocks.go -package=mocks

type (
	notifier interface {
		SendMessageNotificationsKindHandler(ctx context.Context, data []byte) error
	}

	kafkaConsumer interface {
		Poll(timeoutMs int) ckafka.Event
		CommitMessage(m *ckafka.Message) ([]ckafka.TopicPartition, error)
		Assign(partitions []ckafka.TopicPartition) error
		Unassign() error
	}

	kafkaProducer interface {
		Produce(msg *ckafka.Message, deliveryChan chan ckafka.Event) error
	}
)

type eventPayload struct {
	UserID  int64  `json:"user_id"`
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

const (
	sessionTimeoutConsumer = 10000
	deliveryTimeout        = 30000
	timeoutPoll            = 100
)

func RunMainConsumer(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	notifier notifier,
) error {
	brokers := cfg.ConstructKafkaBrokers()
	if len(brokers) == 0 {
		return errors.New("empty kafka brokers")
	}

	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers":        joinBrokers(brokers),
		"group.id":                 cfg.Kafka.ConsumerGroup,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
		"session.timeout.ms":       sessionTimeoutConsumer,
	})
	if err != nil {
		return err
	}

	producer, err := ckafka.NewProducer(&ckafka.ConfigMap{
		"bootstrap.servers":   joinBrokers(brokers),
		"acks":                "-1",
		"retries":             10,
		"delivery.timeout.ms": deliveryTimeout,
	})

	if err != nil {
		logger.Error("create kafka producer failed",
			zap.Error(err),
			zap.String("brokers", joinBrokers(brokers)),
		)
		_ = consumer.Close()
		return fmt.Errorf("create kafka producer: %w", err)
	}

	defer func() {
		producer.Close()
		if err := consumer.Close(); err != nil {
			logger.Error("close kafka consumer", zap.Error(err))
		}
	}()

	c := newMainConsumer(
		logger,
		consumer,
		producer,
		cfg.Kafka.Topic,
		notifier,
	)

	if err := consumer.SubscribeTopics([]string{cfg.Kafka.Topic}, nil); err != nil {
		return err
	}

	logger.Info("kafka consumer started",
		zap.Strings("brokers", brokers),
		zap.String("topic", cfg.Kafka.Topic),
		zap.String("group", cfg.Kafka.ConsumerGroup),
	)

	c.run(ctx)
	return nil
}

type mainConsumer struct {
	logger   *zap.Logger
	consumer kafkaConsumer
	producer kafkaProducer

	topic    string
	dlqTopic string

	notifier notifier
}

func newMainConsumer(
	logger *zap.Logger,
	consumer kafkaConsumer,
	producer kafkaProducer,
	topic string,
	notifier notifier,
) *mainConsumer {
	return &mainConsumer{
		logger:   logger,
		consumer: consumer,
		producer: producer,
		topic:    topic,
		dlqTopic: topic + "-dlq",
		notifier: notifier,
	}
}

func (c *mainConsumer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("kafka consumer stopped", zap.Error(ctx.Err()))
			return
		default:
		}

		ev := c.consumer.Poll(timeoutPoll)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *ckafka.Message:
			c.logger.Info("message fetched",
				zap.String("topic", *e.TopicPartition.Topic),
				zap.Int32("partition", e.TopicPartition.Partition),
				zap.Int64("offset", int64(e.TopicPartition.Offset)),
			)

			var payload eventPayload
			if err := json.Unmarshal(e.Value, &payload); err != nil {
				if dlqErr := c.produceDLQ(ctx, e, err); dlqErr != nil {
					c.logger.Error("add message to dlq", zap.Error(dlqErr))
					time.Sleep(time.Second)
				}
				continue
			}

			if err := c.notifier.SendMessageNotificationsKindHandler(ctx, e.Value); err != nil {
				c.logger.Error("failed to send HTTP notification from kafka event",
					zap.Error(err),
					zap.Int64("order_id", payload.OrderID),
				)
				time.Sleep(time.Second)
				continue
			}

			_, err := c.consumer.CommitMessage(e)
			if err != nil {
				c.logger.Error("kafka commit failed", zap.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}

			c.logger.Info("message committed",
				zap.String("topic", *e.TopicPartition.Topic),
				zap.Int64("offset", int64(e.TopicPartition.Offset)),
			)

		case ckafka.Error:
			c.logger.Warn("kafka poll error", zap.Error(e))
			time.Sleep(1 * time.Second)

		case ckafka.AssignedPartitions:
			c.logger.Info("kafka partitions assigned", zap.Any("partitions", e.Partitions))
			if err := c.consumer.Assign(e.Partitions); err != nil {
				c.logger.Error("kafka assign partitions", zap.Error(err))
			}

		case ckafka.RevokedPartitions:
			c.logger.Info("kafka partitions revoked", zap.Any("partitions", e.Partitions))
			if err := c.consumer.Unassign(); err != nil {
				c.logger.Error("kafka unassign partitions", zap.Error(err))
			}
		default:
		}
	}
}

func (c *mainConsumer) produceMessage(ctx context.Context, topic string, key, value []byte) error {
	deliveryChan := make(chan ckafka.Event)
	defer close(deliveryChan)

	err := c.producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: ckafka.PartitionAny,
		},
		Key:   key,
		Value: value,
	}, deliveryChan)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case event := <-deliveryChan:
		m := event.(*ckafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
		}
	}
	return nil
}

func (c *mainConsumer) produceDLQ(ctx context.Context, msg *ckafka.Message, reason error) error {
	c.logger.Error("kafka message unmarshal failed, redirecting to DLQ", zap.Error(reason))

	dlqPayload, err := json.Marshal(map[string]any{
		"original_topic":     *msg.TopicPartition.Topic,
		"original_partition": msg.TopicPartition.Partition,
		"original_offset":    int64(msg.TopicPartition.Offset),
		"original_key":       string(msg.Key),
		"original_value":     string(msg.Value),
		"error":              reason.Error(),
		"failed_at":          time.Now(),
	})
	if err != nil {
		return fmt.Errorf("marshal dlq payload: %w", err)
	}

	if err := c.produceMessage(ctx, c.dlqTopic, msg.Key, dlqPayload); err != nil {
		return fmt.Errorf("produce message to dlq: %w", err)
	}

	if _, err := c.consumer.CommitMessage(msg); err != nil {
		return fmt.Errorf("commit invalid kafka message: %w", err)
	}

	return nil
}

func joinBrokers(brokers []string) string {
	return strings.Join(brokers, ",")
}
