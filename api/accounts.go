package api

import (
	"net/http"

	db "github.com/bank_go/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

func (server *Server) CreateAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	param := db.CreateAccountParams{Owner: req.Owner, Balance: 0, Currency: req.Currency}

	acc, err := server.store.CreateAccount(ctx, param)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, acc)

}

type getAccountRequest struct {
	ID int64 `uri:"id" bindings:"required,min=1"`
}

func (server *Server) GetAccount(ctx *gin.Context) {
	var req getAccountRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	acc, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, acc)

}

type getAccountListReq struct {
	Page    int32 `form:"page" bindings:"required,min=1"`
	PerPage int32 `form:"per_page" bindings:"required,min=10,max=25"`
}

func (server *Server) GetAccountList(ctx *gin.Context) {
	var req getAccountListReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	param := db.ListAccountsParams{Limit: req.PerPage, Offset: (req.Page - 1) * req.PerPage}

	accounts, err := server.store.ListAccounts(ctx, param)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, accounts)

}
