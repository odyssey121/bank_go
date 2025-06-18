package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func createUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) CreateUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPass, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	param := db.CreateUserParams{Username: req.Username, FullName: req.FullName, Email: req.Email, HashedPassword: hashedPass}
	user, err := server.store.CreateUser(ctx, param)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			fmt.Println("pq err = ", err)
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}

		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := createUserResponse(user)
	ctx.JSON(http.StatusOK, resp)
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"required"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserResponse struct {
	User userResponse `json:"user"`
}

var ErrConnotUpdateThisUser = errors.New("auth user cannot update incoming username")

func (server *Server) UpdateUser(ctx *gin.Context) {
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authSubject := ctx.MustGet(authorizationContextKey)

	if req.Username != authSubject {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrConnotUpdateThisUser))
		return

	}

	param := db.UpdateUserParams{
		Username: req.Username,
		FullName: sql.NullString{String: req.FullName, Valid: req.FullName != ""},
		Email:    sql.NullString{String: req.Email, Valid: req.Email != ""},
	}

	if req.Password != "" {
		hashedPassword, err := util.HashPassword(req.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		param.HashedPassword = sql.NullString{String: hashedPassword, Valid: true}
		param.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: true}

	}

	user, err := server.store.UpdateUser(ctx, param)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}

		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, UpdateUserResponse{User: createUserResponse(user)})
}

type loginUserReq struct {
	Username string `json:"username"`
	Passowrd string `json:"password"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"token"`
	AccessTokenExpiresAt  time.Time    `json:"expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) LoginUser(ctx *gin.Context) {
	var req loginUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPasswordHash(req.Passowrd, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, payload, err := server.tokenMaker.CreateToken(user.Username, server.cfg.JwtTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(user.Username, server.cfg.JWtRefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	param := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		RefreshToken: refreshToken,
		Username:     req.Username,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	session, err := server.store.CreateSession(ctx, param)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  payload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  createUserResponse(user),
	}
	ctx.JSON(http.StatusOK, resp)
}

type getUserRequest struct {
	Username string `uri:"username" binding:"required"`
}

var (
	ErrUserDoNotToBelongAuthUser = errors.New("cannot get this user, doesn't belong to the authenticated user")
)

func (server *Server) GetUser(ctx *gin.Context) {
	var req getUserRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	authSubject := ctx.MustGet(authorizationContextKey)

	if user.Username != authSubject {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrUserDoNotToBelongAuthUser))
		return
	}

	ctx.JSON(http.StatusOK, user)

}
