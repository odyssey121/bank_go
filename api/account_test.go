package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/bank_go/db/mock"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func GetRandomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(666, 999),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func TestListAccountApi(t *testing.T) {
	type Query struct {
		page    int32
		perPage int32
	}
	user := GetRandomUser(t)
	n := 10
	accounts := make([]db.Account, n)
	for i := range n {
		accounts[i] = GetRandomAccount(user.Username)
	}
	testCases := []struct {
		name       string
		query      Query
		setupAuth  func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: Query{page: 1, perPage: 10},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccountsByOwner(gomock.Any(), gomock.Any()).Times(1).Return(accounts, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respAccounts []db.Account
				var err error
				require.Equal(t, http.StatusOK, recorder.Code)
				err = json.Unmarshal(recorder.Body.Bytes(), &respAccounts)
				require.NoError(t, err)
				require.True(t, len(respAccounts) == n)
				require.Equal(t, accounts, respAccounts)
			},
		},
		{
			name:  "NoAuthorization",
			query: Query{page: 1, perPage: 10},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {

			},
			buildStubs: func(store *mockdb.MockStore) {

			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:  "InternalServerError",
			query: Query{page: 1, perPage: 10},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccountsByOwner(gomock.Any(), gomock.Any()).Times(1).Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "IncorrectPageSize",
			query: Query{page: 1, perPage: 99},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccountsByOwner(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
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

			request, err := http.NewRequest(http.MethodGet, "/accounts", nil)
			require.NoError(t, err)
			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page", fmt.Sprintf("%d", tc.query.page))
			q.Add("per_page", fmt.Sprintf("%d", tc.query.perPage))
			request.URL.RawQuery = q.Encode()
			// add header token
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})
	}

}

func TestGetAccountApi(t *testing.T) {
	user := GetRandomUser(t)
	account := GetRandomAccount(user.Username)
	testCases := []struct {
		name       string
		accID      int64
		setupAuth  func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			accID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respAccount db.Account
				var err error
				require.Equal(t, http.StatusOK, recorder.Code)
				err = json.Unmarshal(recorder.Body.Bytes(), &respAccount)
				require.NoError(t, err)
				require.EqualValues(t, account, respAccount)
			},
		},
		{
			name:  "NotFound",
			accID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respAccount db.Account
				var err error
				require.Equal(t, http.StatusNotFound, recorder.Code)
				err = json.Unmarshal(recorder.Body.Bytes(), &respAccount)
				require.NoError(t, err)
				require.Empty(t, respAccount)
			},
		},
		{
			name:  "InternalServerError",
			accID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				var acc db.Account
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(acc, sql.ErrConnDone)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respAccount db.Account
				var err error
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				err = json.Unmarshal(recorder.Body.Bytes(), &respAccount)
				require.NoError(t, err)
				require.Empty(t, respAccount)
			},
		},
		{
			name:  "InvalidID",
			accID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respAccount db.Account
				var err error
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				err = json.Unmarshal(recorder.Body.Bytes(), &respAccount)
				require.NoError(t, err)
				require.Empty(t, respAccount)
			},
		},
		{
			name:  "DoNotBelongAuthUser",
			accID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, "anonimus", request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var errResp gin.H
				util.DeSerializeGinErr(&errResp, recorder.Result().Body)
				require.Equal(t, errResp["error"], ErrAccountDoNotToBelongAuthUser.Error())
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

			url := fmt.Sprintf("/accounts/%d", tc.accID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// add header token
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})
	}
}

func TestCreateAccountApi(t *testing.T) {
	user := GetRandomUser(t)
	account := GetRandomAccount(user.Username)
	testCases := []struct {
		name       string
		body       gin.H
		setupAuth  func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, user.Username, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				param := db.CreateAccountParams{Owner: user.Username, Currency: account.Currency, Balance: int64(0)}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(param)).Times(1).Return(account, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respAccount db.Account
				var err error
				require.Equal(t, http.StatusOK, recorder.Code)
				err = json.Unmarshal(recorder.Body.Bytes(), &respAccount)
				require.NoError(t, err)
				require.Equal(t, user.Username, respAccount.Owner)
				require.EqualValues(t, account, respAccount)
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

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(data))
			require.NoError(t, err)

			// add header token
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})
	}

}
