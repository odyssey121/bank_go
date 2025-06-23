package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bank_go/api"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/gapi"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/queues"
	"github.com/bank_go/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	var err error
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	// logger
	if cfg.Enviroment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(cfg.DbDriver, cfg.DbConnectionString)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	store := db.NewStore(conn)
	// asyncq init
	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisServerAddress}
	qtProvider := queues.NewRedisProvider(redisOpt)

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	waitGroup, ctx := errgroup.WithContext(ctx)

	runGrpcServer(ctx, waitGroup, store, cfg, qtProvider)
	runGrpcGatewayServer(ctx, waitGroup, cfg, store, qtProvider)
	runAsyncQServer(ctx, waitGroup, store, redisOpt)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func runAsyncQServer(ctx context.Context, wg *errgroup.Group, store db.Store, opt asynq.RedisClientOpt) {
	queueHandler := queues.NewRedisTaskHandler(store, opt)

	wg.Go(func() error {
		err := queueHandler.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot start asyncq server")
		}
		return err
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown task processor")

		queueHandler.Shutdown()
		log.Info().Msg("task processor is stopped")

		return nil
	})

}

func runGrpcServer(ctx context.Context, wg *errgroup.Group, store db.Store, cfg util.Config, qtProvider queues.TaskProvider) {
	gapiServer, err := gapi.NewServer(store, cfg, qtProvider)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create gapi server")
	}

	lis, err := net.Listen("tcp", cfg.GrpcServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	// logger
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	reflection.Register(grpcServer)
	pb_sources.RegisterBankGoServer(grpcServer, gapiServer)

	wg.Go(func() error {
		log.Info().Msgf("gapi server listening at %v", lis.Addr())

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed to gapi serve")
			return err
		}

		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown gRPC server")

		grpcServer.GracefulStop()
		log.Info().Msg("gRPC server is stopped")

		return nil
	})

}

func runGrpcGatewayServer(ctx context.Context, wg *errgroup.Group, cfg util.Config, store db.Store, qtProvider queues.TaskProvider) {
	server, err := gapi.NewServer(store, cfg, qtProvider)
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	mux := runtime.NewServeMux(jsonOption)

	handler := gapi.HttpLogger(mux)

	err = pb_sources.RegisterBankGoHandlerServer(ctx, mux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to RegisterBankGoHandlerFromEndpoint")
	}

	httpServer := &http.Server{
		Handler: handler,
		Addr:    cfg.WebServerAddress,
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)

	wg.Go(func() error {
		log.Info().Msgf("gapi gateway server listening at %s", cfg.WebServerAddress)
		err = httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("HTTP gateway server failed to serve")
			return err
		}
		return err
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("graceful shutdown HTTP gateway server")

		err := httpServer.Shutdown(context.Background())

		if err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP gateway server")
			return err
		}

		log.Info().Msg("HTTP gateway server is stopped")
		return nil
	})

}

func runGinServer(store db.Store, cfg util.Config) {
	server, err := api.NewServer(store, cfg)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create to new server")
	}

	err = server.Start(cfg.WebServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
