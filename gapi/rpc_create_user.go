package gapi

import (
	"context"

	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb_sources.CreateUserRequest) (*pb_sources.CreateUserResponse, error) {
	hashedPass, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	param := db.CreateUserParams{Username: req.GetUsername(), FullName: req.GetFullName(), Email: req.GetEmail(), HashedPassword: hashedPass}
	user, err := server.store.CreateUser(ctx, param)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "%s", err.Error())
			}
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	resp := &pb_sources.CreateUserResponse{User: convertUser(user)}
	return resp, nil
}
