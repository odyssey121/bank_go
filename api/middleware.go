package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bank_go/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationHeaderType = "bearer"
	authorizationContextKey = "authorization_subject"
)

func authMiddleware(tokenMaker token.Maker[*token.PasetoPayload]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		headerAuthVal := ctx.GetHeader(authorizationHeaderKey)
		if len(headerAuthVal) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		headerAuthFields := strings.Fields(headerAuthVal)
		if len(headerAuthFields) < 2 {
			err := errors.New("incorrect authorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		if strings.ToLower(headerAuthFields[0]) != authorizationHeaderType {
			err := errors.New("incorrect authorization header type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		verfyResponse, err := tokenMaker.VerifyToken(headerAuthFields[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationContextKey, verfyResponse.Username)
		ctx.Next()

	}
}
