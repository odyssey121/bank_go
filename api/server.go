package api

import (
	"github.com/bank_go/token"
	"github.com/bank_go/util"

	db "github.com/bank_go/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker[*token.PasetoPayload]
	cfg        util.Config
}

func NewServer(store db.Store, cfg util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.JwtSecretKey)
	if err != nil {
		return nil, err
	}

	server := &Server{store: store, tokenMaker: tokenMaker, cfg: cfg}
	server.setupRoulter()
	return server, nil
}

func (server *Server) setupRoulter() {
	r := gin.Default()

	r.POST("/users/login", server.LoginUser)

	authorizedGroup := r.Group("/")
	authorizedGroup.Use(authMiddleware(server.tokenMaker))
	authorizedGroup.POST("/accounts", server.CreateAccount)
	authorizedGroup.GET("/accounts/:id", server.GetAccount)
	authorizedGroup.GET("/accounts", server.GetAccountList)
	authorizedGroup.POST("/transfer", server.Transfer)
	r.POST("/users", server.CreateUser)
	authorizedGroup.GET("/users/:username", server.GetUser)

	server.router = r
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}

}
