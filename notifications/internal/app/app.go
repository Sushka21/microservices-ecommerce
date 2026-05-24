package app

import (
	"context"
	"errors"
	"net"
	"os/signal"
	"syscall"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/consumer/kafka"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/controller"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/inbox"
	repoInbox "github.com/Sushka21/microservices-ecommerce/notifications/internal/repository/inbox"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/repository/transactor"
	notificationUsecase "github.com/Sushka21/microservices-ecommerce/notifications/internal/usecase/notifications"
	db "github.com/Sushka21/microservices-ecommerce/notifications/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var ErrKafkaBrokersNotConfigured = errors.New("no kafka brokers configured (KAFKA_BROKERS)")

func Run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	notificationsClient := notificationUsecase.NewNotificationsService(logger, cfg)

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

	inboxRepo := repoInbox.NewInboxRepository(dbpool)

	kafkaBrokers := cfg.ConstructKafkaBrokers()

	if len(kafkaBrokers) == 0 {
		logger.Error("no kafka brokers configured (KAFKA_BROKERS)")
		return ErrKafkaBrokersNotConfigured
	}

	globalInboxHandler := func(kind repoInbox.Kind) (inbox.KindHandler, error) {
		switch kind {
		case repoInbox.KindNotification:
			return notificationsClient.SendMessageNotificationsKindHandler, nil
		default:
			return nil, errors.New("unsupported outboxCore kind")
		}
	}

	inboxcore := inbox.New(logger, inboxRepo, globalInboxHandler, cfg, transactor)

	inboxcore.Start(
		ctx,
		cfg.Inbox.Workers,
		cfg.Inbox.BatchSize,
		cfg.Inbox.FetchPeriod,
		cfg.Inbox.TTL,
	)

	ctrl := controller.New(notificationsClient, logger)

	wgerr, ctx := errgroup.WithContext(ctx)

	wgerr.Go(func() error {
		return kafka.RunMainConsumer(ctx, logger, cfg, inboxRepo, transactor)
	})

	wgerr.Go(func() error {
		return RunGrpc(ctx, logger, cfg, ctrl)
	})

	if err := wgerr.Wait(); err != nil {
		return err
	}
	return nil
}

func RunGrpc(ctx context.Context, logger *zap.Logger, cfg *config.Config, ctrl *controller.API) error {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("can not open tcp socket", zap.Error(err))
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
