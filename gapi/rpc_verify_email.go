package gapi

import (
	"context"

	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb_sources.VerifyEmailRequest) (*pb_sources.VerifyEmailResponse, error) {
	val := validateVerifyEmailRequest(req)
	if val != nil {
		return nil, invalidArgumentError(val)
	}

	txResult, err := server.store.EmailVerifyTx(ctx, db.EmailVerifyTxParam{Id: req.GetId(), Code: req.GetCode()})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}

	res := &pb_sources.VerifyEmailResponse{IsVerified: txResult.User.IsEmailVerified.Bool}

	return res, nil
}

func validateVerifyEmailRequest(req *pb_sources.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	return violations
}
