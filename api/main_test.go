package api

import (
	"os"
	"testing"
	"time"

	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store) *Server {
	cfg := util.Config{JwtSecretKey: util.RandomString(32), JwtTokenDuration: time.Minute}
	tokenMaker, err := token.NewPasetoMaker(cfg.JwtSecretKey)
	require.NoError(t, err)
	server := &Server{store: store, tokenMaker: tokenMaker, cfg: cfg}
	server.setupRoulter()
	return server

}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
