package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	mockdb "github.com/bank_go/db/mock"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userParamEqMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e userParamEqMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPasswordHash(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e userParamEqMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", e.arg, e.arg)
}

func eqCreateUserParam(arg db.CreateUserParams, password string) gomock.Matcher {
	return userParamEqMatcher{arg, password}
}

func GetRandomUser(t *testing.T) db.User {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	return db.User{
		Username:       util.RandomString(6),
		FullName:       util.RandomString(16),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
	}
}

func TestCreateUserApi(t *testing.T) {
	user := GetRandomUser(t)
	password := util.RandomString(6)
	hashedPass, err := util.HashPassword(password)
	require.NoError(t, err)
	// fmt.Println("hashedPass:", hashedPass)
	user.HashedPassword = hashedPass
	// fmt.Println("user.HashedPassword:", user.HashedPassword)

	testCases := []struct {
		name           string
		reqUnserialize createUserRequest
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs     func(store *mockdb.MockStore)
		checkResp      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			reqUnserialize: createUserRequest{Username: user.Username, FullName: user.FullName, Email: user.Email, Password: password},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				param := db.CreateUserParams{Username: user.Username, FullName: user.FullName, Email: user.Email, HashedPassword: user.HashedPassword}
				store.EXPECT().CreateUser(gomock.Any(), eqCreateUserParam(param, password)).Times(1).Return(user, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respUser userResponse
				require.Equal(t, http.StatusOK, recorder.Code)
				err := json.Unmarshal(recorder.Body.Bytes(), &respUser)
				require.NoError(t, err)
				require.Equal(t, respUser.Username, user.Username)
				require.Equal(t, respUser.FullName, user.FullName)
				require.Equal(t, respUser.Email, user.Email)
				require.Equal(t, respUser.PasswordChangedAt, user.PasswordChangedAt)
			},
		},
		{
			name:           "BadRequestValidation",
			reqUnserialize: createUserRequest{Username: user.Username, Email: "user.Email", Password: "1"},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var errResp gin.H
				util.DeSerializeGinErr(&errResp, recorder.Result().Body)
				assert.Contains(t, errResp["error"],
					"'createUserRequest.FullName' Error:Field validation for 'FullName' failed on the 'required' tag",
					"'createUserRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag",
					"'createUserRequest.Email' Error:Field validation for 'Email' failed on the 'email'",
				)

				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// test build stubs
			tc.buildStubs(store)

			// start test server request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			payloadJson := new(bytes.Buffer)
			util.Serialize(tc.reqUnserialize, payloadJson)
			// fmt.Println("payloadJson:", payloadJson)

			request, err := http.NewRequest(http.MethodPost, "/users", payloadJson)
			require.NoError(t, err)
			//add auth header
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})

	}

}

func TestGetUserApi(t *testing.T) {
	user := GetRandomUser(t)
	testCases := []struct {
		name       string
		username   string
		setupAuth  func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(user, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respUser db.User
				require.Equal(t, http.StatusOK, recorder.Code)
				err := json.Unmarshal(recorder.Body.Bytes(), &respUser)
				require.NoError(t, err)
				require.EqualValues(t, user, respUser)
			},
		}, {
			name:     "DoNotBelongAuthUser",
			username: user.Username,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, "anonimus", request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(user, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var errResp gin.H
				util.DeSerializeGinErr(&errResp, recorder.Result().Body)
				require.Equal(t, errResp["error"], ErrUserDoNotToBelongAuthUser.Error())
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:     "NotFound",
			username: "not_exist",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Eq("not_exist")).Times(1).Return(db.User{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// test build stubs
			tc.buildStubs(store)

			// start test server request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			// add header authorization
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)
		})

	}

}
