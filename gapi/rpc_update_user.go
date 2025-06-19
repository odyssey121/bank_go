package gapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb_sources.UpdateUserRequest) (*pb_sources.UpdateUserResponse, error) {

	param := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{String: req.GetFullName(), Valid: len(req.FullName) > 0},
		Email:    sql.NullString{String: req.GetEmail(), Valid: len(req.Email) >= 4},
	}

	if req.Username == "" {
		return nil, status.Errorf(codes.Canceled, "username not given")
	}

	if len(req.Password) >= 6 {
		hashedPass, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password")
		}
		param.HashedPassword = sql.NullString{String: hashedPass, Valid: true}
		param.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	user, err := server.store.UpdateUser(ctx, param)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user with username '%s' not found", req.GetUsername())

		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err.Error())
	}

	resp := &pb_sources.UpdateUserResponse{User: convertUser(user)}
	return resp, nil
}
