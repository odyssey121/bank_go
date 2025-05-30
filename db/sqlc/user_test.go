package db

import (
	"context"
	"testing"
	"time"

	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	user1 := User{Username: util.RandomOwner(), HashedPassword: "secret", FullName: util.RandomOwner(), Email: util.RandomEmail()}
	param := CreateUserParams{Username: user1.Username, FullName: user1.FullName, Email: user1.Email, HashedPassword: user1.HashedPassword}
	user2, err := sqlcQueries.CreateUser(context.Background(), param)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Email, user2.Email)
	require.True(t, user2.PasswordChangedAt.IsZero())
	require.NotZero(t, user2.CreatedAt)
	return user2
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)

}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := sqlcQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1, user1)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}
