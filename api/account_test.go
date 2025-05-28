package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/bank_go/db/mock"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func GetRandomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(666, 999),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func TestGetAccountApi(t *testing.T) {
	account := GetRandomAccount()
	testCases := []struct {
		name       string
		accID      int64
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			accID: account.ID,
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
			buildStubs: func(store *mockdb.MockStore) {
				var acc db.Account
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(acc, sql.ErrNoRows)
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
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})

	}

}
