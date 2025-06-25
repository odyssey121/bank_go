package db

import (
	"context"
)

type CreateUserTxParam struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, p CreateUserTxParam) (CreateUserTxResult, error) {
	var result CreateUserTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, CreateUserParams{Username: p.Username, Email: p.Email, FullName: p.FullName, HashedPassword: p.HashedPassword})
		if err != nil {
			return err
		}
		return p.AfterCreate(result.User)
	})
	return result, err

}
