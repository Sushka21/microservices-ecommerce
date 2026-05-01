package config

import (
	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		GRPC struct {
			Port string `env:"GRPC_PORT" envDefault:"50053"`
		}
		Clients struct {
			CallbackAddr string `env:"CALLBACK_ADDR" envDefault:""`
		}
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}



