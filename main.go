package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/bank_go/api"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/gapi"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	var err error
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	// loger
	if cfg.Enviroment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(cfg.DbDriver, cfg.DbConnectionString)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	store := db.NewStore(conn)

	go runGrpcGatewayServer(cfg)
	runGrpcServer(store, cfg)

}

func runGrpcServer(store db.Store, cfg util.Config) {
	gapiServer, err := gapi.NewServer(store, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create gapi server")
	}

	lis, err := net.Listen("tcp", cfg.GrpcServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	// loger
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	reflection.Register(grpcServer)
	pb_sources.RegisterBankGoServer(grpcServer, gapiServer)

	log.Info().Msgf("gapi server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to gapi serve")
	}

}

func runGrpcGatewayServer(cfg util.Config) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	mux := runtime.NewServeMux(jsonOption)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := pb_sources.RegisterBankGoHandlerFromEndpoint(ctx, mux, cfg.GrpcServerAddress, opts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to RegisterBankGoHandlerFromEndpoint")
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	log.Info().Msgf("gapi gateway server listening at %s", cfg.WebServerAddress)
	err = http.ListenAndServe(cfg.WebServerAddress, mux)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to gapi gateway serve")
	}

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
