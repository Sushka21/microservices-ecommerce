package config

import (
	"fmt"
	"net"
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
		Clients struct {
			LOMSGrpcAddr string `env:"LOMS_GRPC_ADDR" envDefault:"localhost:50052"`
		}

		GRPC struct {
			Host        string `env:"GRPC_HOST" envDefault:"localhost"`
			Port        string `env:"GRPC_PORT" envDefault:"50051"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8080"`
		}
		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"loms"`
			User     string `env:"POSTGRES_USER" envDefault:"loms"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
		}
	}
)

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



