package db

import (
	"context"
)

type CreateUserTxParam struct {
	CreateUserParam CreateUserParams
	AfterCreate     func(user User) error
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, param CreateUserTxParam) (CreateUserTxResult, error) {
	var result CreateUserTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, param.CreateUserParam)
		if err != nil {
			return err
		}
		return param.AfterCreate(result.User)
	})
	return result, err

}
