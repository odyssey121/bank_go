package gapi

import (
	"testing"
	"time"

	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/queues"
	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store, qtProvider queues.TaskProvider) *Server {
	cfg := util.Config{
		JwtSecretKey:     util.RandomString(32),
		JwtTokenDuration: time.Minute,
	}
	server, err := NewServer(store, cfg, qtProvider)
	require.NoError(t, err)

	return server

}
