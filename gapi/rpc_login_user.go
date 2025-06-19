package gapi

import (
	"context"
	"database/sql"

	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb_sources.LoginUserRequest) (*pb_sources.LoginUserResponse, error) {
	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "%s", err)
	}

	err = util.CheckPasswordHash(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
	}

	accessToken, payload, err := server.tokenMaker.CreateToken(user.Username, server.cfg.JwtTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token error = %s", err)
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(user.Username, server.cfg.JWtRefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token error = %s", err)
	}

	md := server.extractMetadata(ctx)

	param := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		RefreshToken: refreshToken,
		Username:     req.Username,
		UserAgent:    md.UserAgent,
		ClientIp:     md.ClientIP,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	session, err := server.store.CreateSession(ctx, param)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session error = %s", err)
	}

	resp := &pb_sources.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(payload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		User:                  convertUser(user),
	}
	return resp, nil
}
