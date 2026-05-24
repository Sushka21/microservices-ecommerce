package main

import (
	"log"

	// "github.com/Sushka21/microservices-ecommerce/loms/internal/app"
	// "github.com/Sushka21/microservices-ecommerce/loms/internal/config"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/app"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can not initialize logger: %s", err)
	}

	cfg, err := config.New()

	if err != nil {
		log.Fatalf("can not initialize config: %s", err)
	}

	if err := app.Run(logger, cfg); err != nil {
		logger.Fatal("app stopped with error", zap.Error(err))
	}
}
