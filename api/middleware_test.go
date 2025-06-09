package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func addAuthorizationHeader(
	t *testing.T,
	username string,
	request *http.Request,
	authHeaderKey string,
	authHeaderType string,
	tokenMaker token.Maker[*token.PasetoPayload],
	tokenDuration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, tokenDuration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	request.Header.Set(authHeaderKey, fmt.Sprintf("%s %s", authHeaderType, token))

}

func TestAuthMiddleware(t *testing.T) {
	username := util.RandomString(6)
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusOK)

			},
		},
		{
			name: "AuthHeaderNotProvided",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)

			},
		},
		{
			name: "AuthHeaderTypeNotIncorrect",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, username, request, authorizationHeaderKey, "incorrect type", tokenMaker, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)

			},
		},
		{
			name: "AuthHeaderTokenExpired",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusUnauthorized)

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := NewTestServer(t, nil)

			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				})

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})

	}

}
