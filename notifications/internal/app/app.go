package app

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/Sushka21/microservices-ecommerce/notifications/internal/config"
	"github.com/Sushka21/microservices-ecommerce/notifications/internal/controller"
	notificationUsecase "github.com/Sushka21/microservices-ecommerce/notifications/internal/usecase/notifications"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	notificationsClient := notificationUsecase.NewNotificationsService(logger, cfg)
	ctrl := controller.New(notificationsClient, logger)

	wgerr, ctx := errgroup.WithContext(ctx)
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



