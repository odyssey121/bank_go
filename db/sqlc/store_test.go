package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreTransferTx(t *testing.T) {
	store := NewStore(testDb)

	accFromBeforeTx := createRandomAccount(t)
	accToBeforeTx := createRandomAccount(t)

	amount := int64(10)
	n := 4

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			res, err := store.TransferTx(context.Background(), TransferTxParam{FromAccountID: accFromBeforeTx.ID, ToAccountID: accToBeforeTx.ID, Amount: amount})
			errs <- err
			results <- res

		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)
		res := <-results
		require.NotEmpty(t, res.Transfer)
		require.NotEmpty(t, res.FromEntry)
		require.NotEmpty(t, res.ToEntry)
		transfer, err := store.GetTransfer(context.Background(), res.Transfer.ID)
		require.NoError(t, err)
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.ID, res.Transfer.ID)
		require.Equal(t, transfer.FromAccountID, res.Transfer.FromAccountID)
		require.Equal(t, transfer.ToAccountID, res.Transfer.ToAccountID)
		require.Equal(t, transfer.Amount, res.Transfer.Amount)
		fromEntry, err := store.GetEntry(context.Background(), res.FromEntry.ID)
		require.NoError(t, err)
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.AccountID, res.FromEntry.AccountID)
		require.Equal(t, fromEntry.Amount, res.FromEntry.Amount)
		toEntry, err := store.GetEntry(context.Background(), res.ToEntry.ID)
		require.NoError(t, err)
		require.NotEmpty(t, toEntry)
		require.Equal(t, toEntry.AccountID, res.ToEntry.AccountID)
		require.Equal(t, toEntry.Amount, res.ToEntry.Amount)
		// check accounts balance
		accFrom, err := store.GetAccount(context.Background(), res.Transfer.FromAccountID)
		require.NoError(t, err)
		require.NotEmpty(t, accFrom)
		// require.Equal(t, accFromBeforeTx.Balance, accFrom.Balance+amount)
		accTo, err := store.GetAccount(context.Background(), res.Transfer.ToAccountID)
		require.NoError(t, err)
		require.NotEmpty(t, accTo)
		// require.Equal(t, accFromBeforeTx.Balance, accFrom.Balance-amount)

		diffBalanceFrom := accFromBeforeTx.Balance - accFrom.Balance
		diffBalanceTo := accTo.Balance - accToBeforeTx.Balance
		require.Equal(t, diffBalanceFrom, diffBalanceTo)

		require.True(t, diffBalanceTo%amount == 0)

	}

}
