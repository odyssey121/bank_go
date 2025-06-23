package db

import "context"

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

type TxKeyStructCtx struct{}

var txKey = TxKeyStructCtx{}

func (store *SQLStore) TransferTx(ctx context.Context, param TransferTxParam) (TransferTxResult, error) {
	var result TransferTxResult
	// ctxV := ctx.Value(txKey)
	// fmt.Println(ctxV)
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{param.FromAccountID, param.ToAccountID, param.Amount})
		// fmt.Println("CreateTransfer >> ", ctxV)
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{param.FromAccountID, -param.Amount})
		// fmt.Println("CreateEntry1 >> ", ctxV)
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{param.ToAccountID, param.Amount})
		// fmt.Println("CreateEntry2 >> ", ctxV)
		if err != nil {
			return err
		}
		// prevent deadlock
		if param.FromAccountID < param.ToAccountID {
			result.FromAccount, err = q.UpdateAccountBalanceMinus(ctx, UpdateAccountBalanceMinusParams{Amount: param.Amount, ID: param.FromAccountID})
			if err != nil {
				return err
			}
			result.ToAccount, err = q.UpdateAccountBalancePlus(ctx, UpdateAccountBalancePlusParams{Amount: param.Amount, ID: param.ToAccountID})
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, err = q.UpdateAccountBalancePlus(ctx, UpdateAccountBalancePlusParams{Amount: param.Amount, ID: param.ToAccountID})
			if err != nil {
				return err
			}
			result.FromAccount, err = q.UpdateAccountBalanceMinus(ctx, UpdateAccountBalanceMinusParams{Amount: param.Amount, ID: param.FromAccountID})
			if err != nil {
				return err
			}

		}

		return nil
	})

	return result, err

}
