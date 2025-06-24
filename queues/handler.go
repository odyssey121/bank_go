package queues

import (
	"context"

	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/mail"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type TaskHandler interface {
	handleEmailVerifyTask(ctx context.Context, t *asynq.Task) error
	Start() error
	Shutdown()
}

type RedisTaskHandler struct {
	server     *asynq.Server
	store      db.Store
	mailSender mail.EmailSender
}

func NewRedisTaskHandler(store db.Store, mailSender mail.EmailSender, redisAsynqOpt asynq.RedisClientOpt) TaskHandler {
	logger := NewLogger()
	redis.SetLogger(logger)
	srv := asynq.NewServer(
		redisAsynqOpt,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				QueueCritical: 6,
				QueueDefault:  3,
				QueueLow:      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).
					Str("type", TypeSendVerifyEmail).
					Bytes("payload", task.Payload()).
					Msg("processing the queue task")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskHandler{server: srv, store: store, mailSender: mailSender}

}

func (redisTaskHandler *RedisTaskHandler) Start() error {

	// mux maps a type to a handler
	mux := asynq.NewServeMux()

	mux.HandleFunc(TypeSendVerifyEmail, redisTaskHandler.handleEmailVerifyTask)
	// ...register other handlers...

	return redisTaskHandler.server.Run(mux)
}

func (redisTaskHandler *RedisTaskHandler) Shutdown() {
	redisTaskHandler.server.Shutdown()
}
