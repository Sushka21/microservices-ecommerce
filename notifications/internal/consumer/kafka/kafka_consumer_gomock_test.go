package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/consumer/kafka/mocks"
	repository "github.com/Sushka21/microservices-ecommerce/notifications/internal/repository/inbox"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestJoinBrokers(t *testing.T) {
	t.Parallel()

	// Arrange
	tests := []struct {
		name    string
		brokers []string
		want    string
	}{
		{
			name:    "empty brokers",
			brokers: nil,
			want:    "",
		},
		{
			name:    "one broker",
			brokers: []string{"localhost:9092"},
			want:    "localhost:9092",
		},
		{
			name: "multiple brokers",
			brokers: []string{
				"localhost:9092",
				"localhost:9093",
				"localhost:9094",
			},
			want: "localhost:9092,localhost:9093,localhost:9094",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got := joinBrokers(tt.brokers)

			// Assert
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCreateKey(t *testing.T) {
	t.Parallel()

	// Arrange
	orderID := int64(123)
	status := "paid"
	kind := repository.KindNotification

	want := fmt.Sprintf("%d_%s_%v", orderID, status, kind)

	// Act
	got := createKey(orderID, status, kind)

	// Assert
	require.Equal(t, want, got)
}

func TestNewMainConsumer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	logger := zap.NewNop()
	consumer := mocks.NewMockkafkaConsumer(ctrl)
	producer := mocks.NewMockkafkaProducer(ctrl)
	notifier := mocks.NewMocknotifier(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	topic := "orders"

	// Act
	c := newMainConsumer(
		logger,
		consumer,
		producer,
		topic,
		notifier,
		transactor,
	)

	// Assert
	require.NotNil(t, c)
	require.Equal(t, logger, c.logger)
	require.Equal(t, consumer, c.consumer)
	require.Equal(t, producer, c.producer)
	require.Equal(t, notifier, c.notifier)
	require.Equal(t, transactor, c.transactor)
	require.Equal(t, topic, c.topic)
	require.Equal(t, "orders-dlq", c.dlqTopic)
}

func TestMainConsumer_ProduceMessage_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrangeа
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		producer: producer,
	}

	topic := "orders"
	key := []byte("key-1")
	value := []byte(`{"hello":"world"}`)

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		DoAndReturn(func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			require.NotNil(t, msg)
			require.NotNil(t, msg.TopicPartition.Topic)
			require.Equal(t, topic, *msg.TopicPartition.Topic)
			require.Equal(t, key, msg.Key)
			require.Equal(t, value, msg.Value)

			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{
						Topic: msg.TopicPartition.Topic,
					},
				}
			}()

			return nil
		})

	// Act
	err := c.produceMessage(context.Background(), topic, key, value)

	// Assert
	require.NoError(t, err)
}

func TestMainConsumer_ProduceMessage_ProducerError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		producer: producer,
	}

	expectedErr := errors.New("produce failed")

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	// Act
	err := c.produceMessage(
		context.Background(),
		"orders",
		[]byte("key-1"),
		[]byte(`{"hello":"world"}`),
	)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestMainConsumer_ProduceMessage_ContextCanceled_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		producer: producer,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		Return(nil)

	// Act
	err := c.produceMessage(
		ctx,
		"orders",
		[]byte("key-1"),
		[]byte(`{"hello":"world"}`),
	)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

func TestMainConsumer_ProduceMessage_DeliveryError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		producer: producer,
	}

	deliveryErr := errors.New("delivery failed")

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		DoAndReturn(func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{
						Topic: msg.TopicPartition.Topic,
						Error: deliveryErr,
					},
				}
			}()

			return nil
		})

	// Act
	err := c.produceMessage(
		context.Background(),
		"orders",
		[]byte("key-1"),
		[]byte(`{"hello":"world"}`),
	)

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "delivery failed")
}

func TestMainConsumer_ProduceDLQ_Success_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	consumer := mocks.NewMockkafkaConsumer(ctrl)
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		consumer: consumer,
		producer: producer,
		dlqTopic: "orders-dlq",
	}

	topic := "orders"
	msg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
			Offset:    10,
		},
		Key:   []byte("key-1"),
		Value: []byte(`bad-json`),
	}

	reason := errors.New("invalid json")

	var producedMsg *ckafka.Message

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		DoAndReturn(func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			producedMsg = msg

			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{
						Topic: msg.TopicPartition.Topic,
					},
				}
			}()

			return nil
		})

	consumer.EXPECT().
		CommitMessage(msg).
		Return([]ckafka.TopicPartition{}, nil)

	// Act
	err := c.produceDLQ(context.Background(), msg, reason)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, producedMsg)
	require.NotNil(t, producedMsg.TopicPartition.Topic)
	require.Equal(t, "orders-dlq", *producedMsg.TopicPartition.Topic)
	require.Equal(t, msg.Key, producedMsg.Key)

	var dlqPayload map[string]any
	require.NoError(t, json.Unmarshal(producedMsg.Value, &dlqPayload))

	require.Equal(t, "orders", dlqPayload["original_topic"])
	require.EqualValues(t, 1, dlqPayload["original_partition"])
	require.EqualValues(t, 10, dlqPayload["original_offset"])
	require.Equal(t, "key-1", dlqPayload["original_key"])
	require.Equal(t, "bad-json", dlqPayload["original_value"])
	require.Equal(t, "invalid json", dlqPayload["error"])
	require.NotEmpty(t, dlqPayload["failed_at"])
}

func TestMainConsumer_ProduceDLQ_ProduceError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	consumer := mocks.NewMockkafkaConsumer(ctrl)
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		consumer: consumer,
		producer: producer,
		dlqTopic: "orders-dlq",
	}

	topic := "orders"
	msg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
			Offset:    10,
		},
		Key:   []byte("key-1"),
		Value: []byte(`bad-json`),
	}

	expectedErr := errors.New("produce failed")

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	// Act
	err := c.produceDLQ(context.Background(), msg, errors.New("invalid json"))

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "produce message to dlq")
	require.ErrorIs(t, err, expectedErr)
}

func TestMainConsumer_ProduceDLQ_CommitError_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	consumer := mocks.NewMockkafkaConsumer(ctrl)
	producer := mocks.NewMockkafkaProducer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		consumer: consumer,
		producer: producer,
		dlqTopic: "orders-dlq",
	}

	topic := "orders"
	msg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
			Offset:    10,
		},
		Key:   []byte("key-1"),
		Value: []byte(`bad-json`),
	}

	expectedErr := errors.New("commit failed")

	producer.EXPECT().
		Produce(gomock.Any(), gomock.Any()).
		DoAndReturn(func(msg *ckafka.Message, deliveryChan chan ckafka.Event) error {
			go func() {
				deliveryChan <- &ckafka.Message{
					TopicPartition: ckafka.TopicPartition{
						Topic: msg.TopicPartition.Topic,
					},
				}
			}()

			return nil
		})

	consumer.EXPECT().
		CommitMessage(msg).
		Return(nil, expectedErr)

	// Act
	err := c.produceDLQ(context.Background(), msg, errors.New("invalid json"))

	// Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), "commit invalid kafka message")
	require.ErrorIs(t, err, expectedErr)
}

func TestMainConsumer_Run_MessageSuccess_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	consumer := mocks.NewMockkafkaConsumer(ctrl)
	producer := mocks.NewMockkafkaProducer(ctrl)
	notifier := mocks.NewMocknotifier(ctrl)
	transactor := mocks.NewMocktransactor(ctrl)

	c := &mainConsumer{
		logger:     zap.NewNop(),
		consumer:   consumer,
		producer:   producer,
		topic:      "orders",
		dlqTopic:   "orders-dlq",
		notifier:   notifier,
		transactor: transactor,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	topic := "orders"
	msg := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
			Offset:    10,
		},
		Key: []byte("key-1"),
		Value: []byte(`{
			"user_id": 1,
			"order_id": 100,
			"status": "paid"
		}`),
	}

	consumer.EXPECT().
		Poll(timeoutPoll).
		Return(msg)

	transactor.EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	notifier.EXPECT().
		SendMessage(
			gomock.Any(),
			createKey(100, "paid", repository.KindNotification),
			repository.KindNotification,
			msg.Value,
		).
		Return(nil)

	consumer.EXPECT().
		CommitMessage(msg).
		DoAndReturn(func(m *ckafka.Message) ([]ckafka.TopicPartition, error) {
			cancel()
			return []ckafka.TopicPartition{}, nil
		})

	// Act
	c.run(ctx)

	// Assert
	require.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestMainConsumer_Run_AssignedPartitions_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	consumer := mocks.NewMockkafkaConsumer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		consumer: consumer,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	partitions := []ckafka.TopicPartition{
		{
			Partition: 1,
		},
	}

	consumer.EXPECT().
		Poll(timeoutPoll).
		Return(ckafka.AssignedPartitions{
			Partitions: partitions,
		})

	consumer.EXPECT().
		Assign(partitions).
		DoAndReturn(func(partitions []ckafka.TopicPartition) error {
			cancel()
			return nil
		})

	// Act
	c.run(ctx)

	// Assert
	require.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestMainConsumer_Run_RevokedPartitions_Gomock(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	// Arrange
	consumer := mocks.NewMockkafkaConsumer(ctrl)

	c := &mainConsumer{
		logger:   zap.NewNop(),
		consumer: consumer,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	consumer.EXPECT().
		Poll(timeoutPoll).
		Return(ckafka.RevokedPartitions{})

	consumer.EXPECT().
		Unassign().
		DoAndReturn(func() error {
			cancel()
			return nil
		})

	// Act
	c.run(ctx)

	// Assert
	require.ErrorIs(t, ctx.Err(), context.Canceled)
}
