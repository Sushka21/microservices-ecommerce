package config

import (
	"fmt"
	"net"
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
			Host        string `env:"GRPC_HOST" envDefault:"localhost"`
			Port        string `env:"GRPC_PORT" envDefault:"50052"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8081"`
		}

		Clients struct {
			NotificationsGrpcAddr string `env:"NOTIFICATIONS_GRPC_ADDR" envDefault:"localhost:50053"`
		}

		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"loms"`
			User     string `env:"POSTGRES_USER" envDefault:"loms"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
		}

		Outbox struct {
			Workers     int           `env:"OUTBOX_WORKERS" envDefault:"3"`
			BatchSize   int           `env:"OUTBOX_BATCH_SIZE" envDefault:"5"`
			FetchPeriod time.Duration `env:"OUTBOX_FETCH_PERIOD" envDefault:"5s"`
			TTL         time.Duration `env:"OUTBOX_IN_PROGRESS_TTL" envDefault:"60s"`
		}

		Kafka struct {
			Brokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
			Topic   string `env:"KAFKA_NOTIFICATIONS_TOPIC" envDefault:"order_status_notifications"`
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

func (c *Config) ConstructPostgresURL() string {
	hostPort := net.JoinHostPort(c.PG.Host, c.PG.Port)

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		c.PG.User,
		c.PG.Password,
		hostPort,
		c.PG.DB,
	)
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
