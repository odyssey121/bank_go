package queues

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskProvider interface {
	ProvideEmailVerifyTask(ctx context.Context, p *EmailVerifyPayload) error
}

type RedisTaskProvider struct {
	client *asynq.Client
}

func NewRedisProvider(opt asynq.RedisClientOpt) TaskProvider {
	c := asynq.NewClient(opt)
	return &RedisTaskProvider{client: c}
}
