package app

import (
	"context"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/Sushka21/microservices-ecommerce/cart/internal/config"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/controller"
	"github.com/Sushka21/microservices-ecommerce/cart/internal/repository/cart"
	db "github.com/Sushka21/microservices-ecommerce/cart/migrations"
	"github.com/jackc/pgx/v5/pgxpool"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	lomsgrpc "github.com/Sushka21/microservices-ecommerce/cart/internal/adapter/loms/grpc"
	productgrpc "github.com/Sushka21/microservices-ecommerce/cart/internal/adapter/product/grpc"

	cartuscase "github.com/Sushka21/microservices-ecommerce/cart/internal/usecase/cart"
	itemuscase "github.com/Sushka21/microservices-ecommerce/cart/internal/usecase/item"

	cartv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/cart/api/cart/v1"
	lomsv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/loms/v1"
	productv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/product/v1"
	stocksv1 "github.com/Sushka21/microservices-ecommerce/pkg/generated/loms/api/stocks/v1"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

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

	if errSetup := db.SetupPostgres(dbpool, logger); errSetup != nil {
		logger.Error("can not creat migrtions", zap.Error(err))
		return errSetup
	}

	lomsConn, err := grpc.NewClient(
		cfg.Clients.LOMSGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("failed to connect to server", zap.Error(err))
		return err
	}
	defer lomsConn.Close()

	productClient := productgrpc.NewProductClient(
		productv1.NewProductServiceClient(lomsConn),
	)

	lomsClient := lomsgrpc.NewLOMSClient(
		lomsv1.NewLomsClient(lomsConn),
		stocksv1.NewStocksClient(lomsConn),
	)

	repo := cart.NewPostgresRepository(dbpool)
	itemService := itemuscase.NewItemService(repo, productClient, lomsClient)
	cartService := cartuscase.NewCartService(repo, productClient, lomsClient)

	ctrl := controller.New(itemService, cartService, logger)

	wgerr, ctx := errgroup.WithContext(ctx)

	wgerr.Go(func() error {
		return runRest(ctx, logger, cfg)
	})

	wgerr.Go(func() error {
		return runGrpc(ctx, logger, cfg, ctrl)
	})

	if err := wgerr.Wait(); err != nil {
		logger.Error("failed to wait", zap.Error(err))
		return err
	}
	return nil
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) error {
	mux := grpcruntime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := cartv1.RegisterCartHandlerFromEndpoint(ctx, mux, net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port), opts); err != nil {
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

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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



