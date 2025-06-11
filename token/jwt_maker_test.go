package token

import (
	"testing"
	"time"

	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMAkerCreateToken(t *testing.T) {
	username := util.RandomOwner()
	duration := time.Minute
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	token, payload, err := maker.CreateToken(username, duration)
	require.NotEmpty(t, payload)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	resultVerify, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotZero(t, resultVerify.ID)
	require.Equal(t, resultVerify.Username, username)
	require.Equal(t, resultVerify.Subject, username)
	require.WithinDuration(t, resultVerify.IssuedAt.Time, issuedAt, time.Second)
	require.WithinDuration(t, resultVerify.ExpiresAt.Time, expiredAt, time.Second)
}

func TestJWTMakerExpiredToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)
	token, payload, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NotEmpty(t, payload)
	require.NoError(t, err)
	verifyResult, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrTokenExpired.Error())
	require.Nil(t, verifyResult)

}
