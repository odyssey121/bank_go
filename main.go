package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/bank_go/api"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/gapi"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	var err error
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(cfg.DbDriver, cfg.DbConnectionString)

	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(conn)

	go runGrpcGatewayServer(cfg)
	runGrpcServer(store, cfg)

}

func runGrpcServer(store db.Store, cfg util.Config) {
	gapiServer, err := gapi.NewServer(store, cfg)
	if err != nil {
		log.Fatalf("cannot create gapi server: %v", err)
	}

	lis, err := net.Listen("tcp", cfg.GrpcServerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	pb_sources.RegisterBankGoServer(grpcServer, gapiServer)
	log.Printf("gapi server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to gapi serve: %v", err)
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
		log.Fatalf("failed to RegisterBankGoHandlerFromEndpoint: %v", err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	log.Printf("gapi gateway server listening at %s", cfg.WebServerAddress)
	err = http.ListenAndServe(cfg.WebServerAddress, mux)
	if err != nil {
		log.Fatalf("failed to gapi gateway serve: %v", err)
	}

}

func runGinServer(store db.Store, cfg util.Config) {
	server, err := api.NewServer(store, cfg)

	if err != nil {
		log.Fatal("cannot create to new server: ", err)
	}

	err = server.Start(cfg.WebServerAddress)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
