package queues

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// A list of task types.
const (
	TypeSendVerifyEmail = "email:sendVerify"
)

type EmailVerifyPayload struct {
	Username string
	Email    string
}

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func (redisTaskProvider *RedisTaskProvider) ProvideEmailVerifyTask(ctx context.Context, p *EmailVerifyPayload, asynqOpts ...asynq.Option) error {
	payload, err := json.Marshal(p)
	if err != nil {
		fmt.Errorf("failde marshal task payload: %w", err)
	}

	task := asynq.NewTask(TypeSendVerifyEmail, payload, asynqOpts...)
	info, err := redisTaskProvider.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("type", TypeSendVerifyEmail).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).
		Msg("enqueue task")

	return nil
}

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
// that satisfies asynq.Handler interface. See examples below.
//---------------------------------------------------------------

func (redisTaskHandler *RedisTaskHandler) handleEmailVerifyTask(ctx context.Context, t *asynq.Task) error {
	var p EmailVerifyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Sending Verify Email to User: username=%s and email=%s", p.Username, p.Email)
	// Email delivery code ...
	// redisTaskHandler.mailSender.SendEmail()
	return nil
}
