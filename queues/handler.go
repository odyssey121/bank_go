package queues

import (
	"context"

	db "github.com/bank_go/db/sqlc"
	"github.com/hibiken/asynq"
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
	server *asynq.Server
}

func NewRedisTaskHandler(store db.Store, redisAsynqOpt asynq.RedisClientOpt) TaskHandler {
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
		},
	)

	return &RedisTaskHandler{server: srv}

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
