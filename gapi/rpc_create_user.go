package gapi

import (
	"context"

	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/queues"
	"github.com/bank_go/util"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb_sources.CreateUserRequest) (*pb_sources.CreateUserResponse, error) {
	val := validateCreateUserRequest(req)
	if val != nil {
		return nil, invalidArgumentError(val)
	}

	hashedPass, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	resTx, err := server.store.CreateUserTx(ctx, db.CreateUserTxParam{
		CreateUserParams: db.CreateUserParams{Username: req.GetUsername(), FullName: req.GetFullName(), Email: req.GetEmail(), HashedPassword: hashedPass},
		AfterCreate: func(user db.User) error {
			payload := &queues.EmailVerifyPayload{Username: user.Username, Email: user.Email, FullName: user.FullName}
			err := server.qtProvider.ProvideEmailVerifyTask(ctx, payload, asynq.Queue(queues.QueueCritical))
			if err != nil {
				log.Error().Err(err).Msgf("failed to handle task <%s> with payload <%s>", queues.TypeSendVerifyEmail, payload)
			}
			return err
		},
	})

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "%s", err.Error())
			}
		}
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &pb_sources.CreateUserResponse{User: convertUser(resTx.User)}, nil
}

func validateCreateUserRequest(req *pb_sources.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := util.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.Email != "" {
		if err := util.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
