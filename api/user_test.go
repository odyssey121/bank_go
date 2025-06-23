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
	arg      db.CreateUserTxParam
	password string
}

func (e userParamEqMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserTxParam)
	if !ok {
		return false
	}

	err := util.CheckPasswordHash(e.password, arg.CreateUserParam.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.CreateUserParam.HashedPassword = arg.CreateUserParam.HashedPassword

	return reflect.DeepEqual(e.arg.CreateUserParam, arg.CreateUserParam)
}

func (e userParamEqMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", e.arg, e.arg)
}

func eqCreateUserParam(arg db.CreateUserTxParam, password string) gomock.Matcher {
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
				txParam := db.CreateUserTxParam{CreateUserParam: param, AfterCreate: func(user db.User) error { return nil }}
				store.EXPECT().CreateUserTx(gomock.Any(), eqCreateUserParam(txParam, password)).Times(1).Return(db.CreateUserTxResult{User: user}, nil)
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

func TestUpdateUserApi(t *testing.T) {
	user1 := GetRandomUser(t)
	user2 := GetRandomUser(t)
	newPassword := util.RandomString(6)
	newHashedPassword, err := util.HashPassword(newPassword)
	require.NoError(t, err)

	updatedUser := db.User{Username: user1.Username, FullName: user2.FullName, Email: user2.Email, HashedPassword: newHashedPassword}

	testCases := []struct {
		name           string
		reqUnserialize UpdateUserRequest
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs     func(store *mockdb.MockStore)
		checkResp      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			reqUnserialize: UpdateUserRequest{Username: user1.Username, FullName: user2.FullName, Email: user2.Email, Password: newPassword},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user1.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// todo add eq matcher for params
				// param := db.UpdateUserParams{
				// 	Username:       user1.Username,
				// 	FullName:       sql.NullString{String: user2.FullName, Valid: true},
				// 	Email:          sql.NullString{String: user2.Email, Valid: true},
				// 	HashedPassword: sql.NullString{String: newHashedPassword, Valid: true},
				// }
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(updatedUser, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respUser UpdateUserResponse
				require.Equal(t, http.StatusOK, recorder.Code)
				err := json.Unmarshal(recorder.Body.Bytes(), &respUser)
				require.NoError(t, err)
				require.Equal(t, respUser.User.Username, updatedUser.Username)
				require.Equal(t, respUser.User.FullName, updatedUser.FullName)
				require.Equal(t, respUser.User.Email, updatedUser.Email)
				require.Equal(t, respUser.User.PasswordChangedAt, updatedUser.PasswordChangedAt)
			},
		},
		{
			name:           "UsernameNotGiven",
			reqUnserialize: UpdateUserRequest{FullName: user2.FullName, Email: user2.Email, Password: newPassword},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user1.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var errResp gin.H
				util.DeSerializeGinErr(&errResp, recorder.Result().Body)
				assert.Contains(t, errResp["error"],
					"Key: 'UpdateUserRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag",
				)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:           "AuthUserCannotUpdate",
			reqUnserialize: UpdateUserRequest{Username: "anonimus", FullName: user2.FullName, Email: user2.Email, Password: newPassword},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user1.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var errResp gin.H
				util.DeSerializeGinErr(&errResp, recorder.Result().Body)
				require.Equal(t, errResp["error"], ErrConnotUpdateThisUser.Error())
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

			request, err := http.NewRequest(http.MethodPut, "/users", payloadJson)
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
