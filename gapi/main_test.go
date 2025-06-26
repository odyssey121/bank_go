package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/queues"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
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

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload], username string, duration time.Duration) context.Context {
	token, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	MD := metadata.MD{
		authorizationHeader: []string{fmt.Sprintf("%s %s", authorizationBearer, token)},
	}
	return metadata.NewIncomingContext(context.Background(), MD)

}
