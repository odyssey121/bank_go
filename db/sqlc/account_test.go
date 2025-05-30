package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)
	param := CreateAccountParams{Owner: user.Username, Balance: util.RandomMoney(), Currency: util.RandomCurrency()}
	acc, err := sqlcQueries.CreateAccount(context.Background(), param)
	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, param.Owner, acc.Owner)
	require.Equal(t, param.Balance, acc.Balance)
	require.Equal(t, param.Currency, acc.Currency)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)
	return acc
}

func TestGetAccount(t *testing.T) {
	acc := createRandomAccount(t)
	acc2, err := sqlcQueries.GetAccount(context.Background(), acc.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.Equal(t, acc, acc2)
	require.WithinDuration(t, acc.CreatedAt, acc2.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	acc := createRandomAccount(t)
	param := UpdateAccountParams{Balance: util.RandomMoney(), ID: acc.ID}
	accUpdated, err := sqlcQueries.UpdateAccount(context.Background(), param)
	require.NoError(t, err)
	require.NotEmpty(t, accUpdated)
	require.NotEqual(t, accUpdated.Balance, acc.Balance)
	require.Equal(t, acc.Currency, accUpdated.Currency)
	require.Equal(t, acc.Owner, accUpdated.Owner)
	require.WithinDuration(t, acc.CreatedAt, accUpdated.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	acc := createRandomAccount(t)
	accDeleted, err := sqlcQueries.DeleteAccount(context.Background(), acc.ID)
	require.NoError(t, err)
	require.Equal(t, acc.ID, accDeleted.ID)
	require.Equal(t, acc.Balance, accDeleted.Balance)
	require.Equal(t, acc.Owner, accDeleted.Owner)
	require.Equal(t, acc.Currency, accDeleted.Currency)
	accExist, err := sqlcQueries.GetAccount(context.Background(), acc.ID)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, accExist)

}

func TestListAccount(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	getListAccParam := ListAccountsParams{Limit: 5, Offset: 5}
	accounts, err := sqlcQueries.ListAccounts(context.Background(), getListAccParam)
	require.NoError(t, err)
	require.Len(t, accounts, 5)
	for _, account := range accounts {
		require.NotEmpty(t, account)
	}

}
