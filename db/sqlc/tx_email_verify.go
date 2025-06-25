package db

import (
	"context"
	"database/sql"
)

type EmailVerifyTxParam struct {
	Id   int64
	Code string
}

type EmailVerifyTxResult struct {
	EmailVerify EmailVerify
	User        User
}

func (store *SQLStore) EmailVerifyTx(ctx context.Context, param EmailVerifyTxParam) (EmailVerifyTxResult, error) {
	var result EmailVerifyTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.EmailVerify, err = q.UpdateEmailVerify(ctx, UpdateEmailVerifyParams{Code: param.Code, ID: param.Id})
		if err != nil {
			return err
		}
		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			IsEmailVerified: sql.NullBool{Bool: true, Valid: true},
			Username:        result.EmailVerify.Username,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return result, err
}
