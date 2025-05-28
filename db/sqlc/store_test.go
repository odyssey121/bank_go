package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDb)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	// fmt.Println("BEFORE >> ", acc1.Balance, acc2.Balance)

	amount := int64(10)
	errors := make(chan error)
	n := 10

	for i := 0; i < n; i++ {
		fromAccID := acc1.ID
		toAccID := acc2.ID
		if i%2 == 1 {
			fromAccID = acc2.ID
			toAccID = acc1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParam{FromAccountID: fromAccID, ToAccountID: toAccID, Amount: amount})
			errors <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errors
		require.NoError(t, err)
	}

	updatedAccount1, _ := store.GetAccount(context.Background(), acc1.ID)
	updatedAccount2, _ := store.GetAccount(context.Background(), acc2.ID)
	// fmt.Println("AFTER >> ", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, acc1.Balance, updatedAccount1.Balance)
	require.Equal(t, acc2.Balance, updatedAccount2.Balance)

}

func TestStoreTransferTx(t *testing.T) {
	store := NewStore(testDb)

	accFromBeforeTx := createRandomAccount(t)
	accToBeforeTx := createRandomAccount(t)
	// fmt.Println("BEFORE >> ", accFromBeforeTx.Balance, accToBeforeTx.Balance)

	amount := int64(10)
	n := 5

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx >> %d", i+1)
		ctx := context.WithValue(context.Background(), txKey, txName)
		go func() {
			res, err := store.TransferTx(ctx, TransferTxParam{FromAccountID: accFromBeforeTx.ID, ToAccountID: accToBeforeTx.ID, Amount: amount})
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
		// check accounts
		accFrom := res.FromAccount
		accTo := res.ToAccount
		// fmt.Println("tx >>", accFrom.Balance, accTo.Balance)
		require.NoError(t, err)
		require.NotEmpty(t, accFrom)
		require.Equal(t, accFromBeforeTx.ID, accFrom.ID)

		require.NoError(t, err)
		require.NotEmpty(t, accTo)
		require.Equal(t, accToBeforeTx.ID, accTo.ID)
		// check accounts balance
		diff1 := accFromBeforeTx.Balance - accFrom.Balance
		diff2 := accTo.Balance - accToBeforeTx.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)
		c := int(diff1 / amount)
		require.True(t, c >= 1 && c <= n)
	}

	accFromAfterTx, _ := store.GetAccount(context.Background(), accFromBeforeTx.ID)
	accToAfterTx, _ := store.GetAccount(context.Background(), accToBeforeTx.ID)
	require.Equal(t, accFromAfterTx.Balance, accFromBeforeTx.Balance-(amount*int64(n)))
	require.Equal(t, accToAfterTx.Balance, accToBeforeTx.Balance+(amount*int64(n)))

}
