package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/bank_go/db/sqlc"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required"`
	ToAccountID   int64  `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,oneof=USD EUR CAD RUB"`
}

func (server *Server) Transfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.isValidAccountCurrency(ctx, req.FromAccountID, req.Currency) {
		return
	}

	if !server.isValidAccountCurrency(ctx, req.ToAccountID, req.Currency) {
		return
	}

	param := db.TransferTxParam{FromAccountID: req.FromAccountID, ToAccountID: req.ToAccountID, Amount: req.Amount}

	res, err := server.store.TransferTx(ctx, param)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, res)

}

func (server *Server) isValidAccountCurrency(ctx *gin.Context, id int64, currency string) bool {
	account, err := server.store.GetAccount(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return false
	}

	if account.Currency != currency {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("account id = %d mismatched with currency %s", account.ID, currency)))
		return false
	}

	return true

}
