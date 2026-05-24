package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/config"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/controller"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/outbox"
	"github.com/Sushka21/microservices-ecommerce/loms/internal/repository/transactor"

	db "github.com/Sushka21/microservices-ecommerce/loms/migrations"
	"github.com/jackc/pgx/v5/pgxpool"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	prpductv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"

	repoOrder "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/order"
	repoOutbox "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/outbox"
	repoProduct "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/product"
	repoStocks "github.com/Sushka21/microservices-ecommerce/loms/internal/repository/stocks"

	lomsUscase "github.com/Sushka21/microservices-ecommerce/loms/internal/usercase/loms"
	productUscase "github.com/Sushka21/microservices-ecommerce/loms/internal/usercase/product"
	stocksUscase "github.com/Sushka21/microservices-ecommerce/loms/internal/usercase/stocks"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/adapter/notifications/kafka"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var ErrKafkaBrokersNotConfigured = errors.New("no kafka brokers configured (KAFKA_BROKERS)")

func Run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pgxcfg, err := pgxpool.ParseConfig(cfg.ConstructPostgresURL())

	if err != nil {
		logger.Error("can not create pgxpool cfg", zap.Error(err))
		return err
	}

	pgxcfg.MaxConns = 8
	pgxcfg.MinConns = 1
	pgxcfg.HealthCheckPeriod = config.HealthCheckPeriod
	pgxcfg.MaxConnLifetime = 0
	pgxcfg.MaxConnIdleTime = config.MaxConnIdleTime

	dbpool, err := pgxpool.NewWithConfig(ctx, pgxcfg)

	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		return err
	}

	defer dbpool.Close()

	if errSetup := db.SetupPostgres(dbpool, logger); errSetup != nil {
		return errSetup
	}

	transactor := transactor.NewTransactor(dbpool)

	ordeRepo := repoOrder.NewPostgresRepository(dbpool)
	stocksRepo := repoStocks.NewPostgresRepository(dbpool)
	productRepo := repoProduct.NewPostgresRepository(dbpool)
	outboxRepo := repoOutbox.NewOutboxRepository(dbpool)

	kafkaBrokers := cfg.ConstructKafkaBrokers()

	if len(kafkaBrokers) == 0 {
		logger.Error("no kafka brokers configured (KAFKA_BROKERS)")
		return ErrKafkaBrokersNotConfigured
	}

	notificationsClient := kafka.New(kafkaBrokers, cfg.Kafka.Topic)
	lomsService := lomsUscase.NewLomsService(ordeRepo, stocksRepo, transactor, notificationsClient, outboxRepo)
	productService := productUscase.NewProductService(productRepo)
	stocksService := stocksUscase.NewStocksService(stocksRepo)

	globalOutBoxHandler := func(kind repoOutbox.Kind) (outbox.KindHandler, error) {
		switch kind {
		case repoOutbox.KindNotification:
			return lomsService.OrderStatusChangedNotificationKindHandler, nil
		default:
			return nil, errors.New("unsupported outboxCore kind")
		}
	}

	outboxcore := outbox.New(logger, outboxRepo, globalOutBoxHandler, cfg, transactor)
	outboxcore.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.FetchPeriod,
		cfg.Outbox.TTL,
	)

	ctrl := controller.New(lomsService, productService, stocksService, logger)

	wgerr, ctx := errgroup.WithContext(ctx)

	wgerr.Go(func() error {
		return runRest(ctx, logger, cfg)
	})

	wgerr.Go(func() error {
		return runGrpc(ctx, logger, cfg, ctrl)
	})

	if err := wgerr.Wait(); err != nil {
		return err
	}
	return nil
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) error {
	mux := grpcruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	endpoint := net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)
	if err := lomsv1.RegisterLomsHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Error("grpc-gateway", zap.Error(err))
		return err
	}
	if err := prpductv1.RegisterProductServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Error("grpc-gateway", zap.Error(err))
		return err
	}
	if err := stocksv1.RegisterStocksHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Error("grpc-gateway", zap.Error(err))
		return err
	}

	handler := corsHandler(mux)
	logger.Info("http gateway", zap.String("addr", ":"+cfg.GRPC.GatewayPort))
	srv := &http.Server{
		Addr:    ":" + cfg.GRPC.GatewayPort,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()
		logger.Info("shutting down http gateway")

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("http shutdown error", zap.Error(err))
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("gateway listen error", zap.Error(err))
		return err
	}
	return nil
}

func runGrpc(ctx context.Context, logger *zap.Logger, cfg *config.Config, ctrl *controller.API) error {
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		logger.Error("grpc listen", zap.Error(err))
		return err
	}

	s := grpc.NewServer()
	reflection.Register(s)
	ctrl.Register(s)

	go func() {
		<-ctx.Done()
		logger.Info("shutting down grpc server")
		s.GracefulStop()
	}()

	logger.Info("grpc", zap.String("addr", ":"+cfg.GRPC.Port))

	if err := s.Serve(lis); err != nil {
		logger.Error("grpc serve", zap.Error(err))
		return err
	}
	return nil
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:5173"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
