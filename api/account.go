package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/bank_go/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,oneof=USD EUR CAD RUB"`
}

func (server *Server) CreateAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authSubject := ctx.MustGet(authorizationContextKey).(string)

	param := db.CreateAccountParams{Owner: authSubject, Balance: 0, Currency: req.Currency}

	acc, err := server.store.CreateAccount(ctx, param)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			fmt.Println("pq err = ", err.Error())
			switch pqErr.Code.Name() {
			case "accounts_owner_fkey", "foreign_key_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, acc)

}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

var (
	ErrAccountDoNotToBelongAuthUser = errors.New("cannot get this account, doesn't belong to the authenticated user")
)

func (server *Server) GetAccount(ctx *gin.Context) {
	var req getAccountRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	acc, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	authSubject := ctx.MustGet(authorizationContextKey)

	if authSubject != acc.Owner {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrAccountDoNotToBelongAuthUser))
		return
	}

	ctx.JSON(http.StatusOK, acc)

}

type getAccountListReq struct {
	Page    int32 `form:"page" binding:"required,min=1"`
	PerPage int32 `form:"per_page" binding:"required,min=10,max=25"`
}

func (server *Server) GetAccountList(ctx *gin.Context) {
	var req getAccountListReq

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authSubject := ctx.MustGet(authorizationContextKey).(string)

	param := db.ListAccountsByOwnerParams{Owner: authSubject, Limit: req.PerPage, Offset: (req.Page - 1) * req.PerPage}

	accounts, err := server.store.ListAccountsByOwner(ctx, param)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, accounts)

}
