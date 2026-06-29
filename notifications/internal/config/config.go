package config

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

const (
	ShutdownTimeout   = 5 * time.Second
	HealthCheckPeriod = 30 * time.Second
	MaxConnIdleTime   = 5 * time.Minute
)

type (
	Config struct {
		GRPC struct {
			Port string `env:"GRPC_PORT" envDefault:"50053"`
		}

		Clients struct {
			CallbackAddr string `env:"CALLBACK_ADDR" envDefault:""`
		}

		Kafka struct {
			Brokers       string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
			Topic         string `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
			ConsumerGroup string `env:"KAFKA_CONSUMER_GROUP" envDefault:"notifications"`
		}
	}
)

func (c *Config) ConstructKafkaBrokers() []string {
	addrs := strings.Split(c.Kafka.Brokers, ",")
	result := make([]string, 0, len(addrs))

	for _, addr := range addrs {
		if addr = strings.TrimSpace(addr); addr != "" {
			result = append(result, addr)
		}
	}

	return result
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
