package queues

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/util"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// A list of task types.
const (
	TypeSendVerifyEmail = "email:sendVerify"
)

type EmailVerifyPayload struct {
	Username string
	FullName string
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
	// db logic
	param := db.CreateEmailVerifyParams{Username: p.Username, Email: p.Email, Code: util.RandomString(32)}
	emailVerify, err := redisTaskHandler.store.CreateEmailVerify(ctx, param)
	if err != nil {
		return err
	}
	// Email delivery code ...
	subject := "Welcome to Bank"
	// TODO: replace this URL with an environment variable that points to a front-end page
	verifyUrl := fmt.Sprintf("http://localhost:8888/v1/verify-email?id=%d&code=%s",
		emailVerify.ID, emailVerify.Code)

	content := fmt.Sprintf(`Hello ,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, verifyUrl)
	to := []string{p.Email}

	err = redisTaskHandler.mailSender.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return err
	}

	log.Info().Str("type", t.Type()).Bytes("payload", t.Payload()).
		Str("email", p.Email).Msg("handled task")

	return nil
}
