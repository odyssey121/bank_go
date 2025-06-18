package gapi

import (
	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
)

type Server struct {
	pb_sources.BankGoServer
	store      db.Store
	tokenMaker token.Maker[*token.PasetoPayload]
	cfg        util.Config
}

func NewServer(store db.Store, cfg util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.JwtSecretKey)
	if err != nil {
		return nil, err
	}

	server := &Server{store: store, tokenMaker: tokenMaker, cfg: cfg}
	return server, nil
}
