package gapi

import (
	"context"
	"database/sql"
	"time"

	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb_sources.VerifyEmailRequest) (*pb_sources.VerifyEmailResponse, error) {
	val := validateVerifyEmailRequest(req)
	res := &pb_sources.VerifyEmailResponse{IsVerified: false}
	if val != nil {
		return nil, invalidArgumentError(val)
	}

	emailVerify, err := server.store.GetEmailVerify(ctx, req.Id)
	if err != nil {
		return res, status.Errorf(codes.Internal, "fail to get email verify: %v", err)
	}

	if emailVerify.Code != req.Code {
		return res, status.Error(codes.InvalidArgument, "the code for the friend with the id does not match")
	}

	if time.Now().After(emailVerify.ExpiredAt) {
		return res, status.Error(codes.Canceled, "verification of email has expired")
	}

	emailVerified, err := server.store.UpdateEmailVerify(ctx, db.UpdateEmailVerifyParams{IsVerified: sql.NullBool{Bool: true, Valid: true}, ID: emailVerify.ID})
	if err != nil {
		return res, status.Errorf(codes.Internal, "fail to update email varified: %v", err)
	}

	_, err = server.store.UpdateUser(ctx, db.UpdateUserParams{
		IsEmailVerify: sql.NullBool{Bool: emailVerified.IsVerified, Valid: emailVerified.IsVerified},
		Username:      emailVerified.Username,
	})
	if err != nil {
		return res, status.Errorf(codes.Internal, "fail to update user is email verify: %v", err)
	}

	res.IsVerified = emailVerified.IsVerified

	return res, nil
}

func validateVerifyEmailRequest(req *pb_sources.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	return violations
}
