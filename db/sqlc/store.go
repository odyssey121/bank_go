package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db, Queries: New(db)}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()

}

type TransferTxParam struct {
	FromAccountID int64
	ToAccountID   int64
	Amount        int64
}

type TransferTxResult struct {
	FromAccount Account
	ToAccount   Account
	Transfer    Transfer
	FromEntry   Entry
	ToEntry     Entry
}

func (store *Store) TransferTx(ctx context.Context, param TransferTxParam) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{param.FromAccountID, param.ToAccountID, param.Amount})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{param.FromAccountID, -param.Amount})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{param.ToAccountID, param.Amount})
		if err != nil {
			return err
		}

		// TODO: update accounts balance

		return nil
	})

	return result, err

}
