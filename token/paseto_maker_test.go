package token

import (
	"testing"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
)

func TestPasetoMakerCreateToken(t *testing.T) {
	username := util.RandomOwner()
	duration := time.Minute
	maker, err := NewPasetoMaker(util.RandomString(chacha20poly1305.KeySize))
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
	require.WithinDuration(t, resultVerify.IssuedAt, issuedAt, time.Second)
	require.WithinDuration(t, resultVerify.ExpiredAt, expiredAt, time.Second)
}

func TestPasetoExpiredToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(chacha20poly1305.KeySize))
	require.NoError(t, err)
	token, payload, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NotEmpty(t, payload)
	require.NoError(t, err)
	verifyResult, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrPasetoExpiredToken.Error())
	require.Nil(t, verifyResult)

}
