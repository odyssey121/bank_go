package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func generatedRandomPassHash(t *testing.T) (password string, hash string) {
	password = RandomString(6)
	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, password, hash)
	return password, hash
}

func TestPasswordHash(t *testing.T) {
	password, hash := generatedRandomPassHash(t)
	wrongPassword := RandomString(6)
	err := CheckPasswordHash(wrongPassword, hash)
	require.Error(t, err)
	err = CheckPasswordHash(password, hash)
	require.NoError(t, err)

}
