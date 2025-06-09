package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/bank_go/db/mock"
	db "github.com/bank_go/db/sqlc"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func getAccountForTransfer(currency string) db.Account {
	return db.Account{ID: util.RandomInt(666, 999), Owner: util.RandomOwner(), Currency: currency, Balance: util.RandomMoney()}
}

func getTransferEntry(account db.Account, amount int64) db.Entry {
	return db.Entry{ID: util.RandomInt(666, 999), AccountID: account.ID, Amount: amount}
}

func TestTransfer(t *testing.T) {
	currency := util.RandomCurrency()
	accountFrom := getAccountForTransfer(currency)
	accountTo := getAccountForTransfer(currency)
	Amount := util.RandomMoney()
	transfer := db.Transfer{ID: util.RandomInt(6, 99), FromAccountID: accountFrom.ID, ToAccountID: accountTo.ID, Amount: Amount}
	fromEntry := getTransferEntry(accountFrom, -Amount)
	toEntry := getTransferEntry(accountTo, Amount)

	testCases := []struct {
		name       string
		req        transferRequest
		txParam    db.TransferTxParam
		txRes      db.TransferTxResult
		setupAuth  func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload])
		buildStubs func(store *mockdb.MockStore, param db.TransferTxParam, txRes db.TransferTxResult)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			txParam: db.TransferTxParam{FromAccountID: accountFrom.ID, ToAccountID: accountTo.ID, Amount: Amount},
			req:     transferRequest{accountFrom.ID, accountTo.ID, Amount, currency},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, accountFrom.Owner, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			txRes: db.TransferTxResult{FromAccount: accountFrom, ToAccount: accountTo, Transfer: transfer, FromEntry: fromEntry, ToEntry: toEntry},
			buildStubs: func(store *mockdb.MockStore, param db.TransferTxParam, txRes db.TransferTxResult) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(accountFrom.ID)).Times(1).Return(accountFrom, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(accountTo.ID)).Times(1).Return(accountTo, nil)
				store.EXPECT().TransferTx(gomock.Any(), param).Times(1).Return(txRes, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var res db.TransferTxResult
				require.Equal(t, http.StatusOK, recorder.Code)
				err := json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)
				require.NotEmpty(t, res)
				require.Equal(t, res.FromAccount.ID, accountFrom.ID)
				require.Equal(t, res.ToAccount.ID, accountTo.ID)
				require.Equal(t, res.Transfer.Amount, Amount)
			},
		},
		{
			name:    "MismatchedCurrency",
			txParam: db.TransferTxParam{FromAccountID: accountFrom.ID, ToAccountID: accountTo.ID, Amount: Amount},
			req:     transferRequest{accountFrom.ID, accountTo.ID, Amount, currency},
			txRes:   db.TransferTxResult{FromAccount: accountFrom, ToAccount: accountTo, Transfer: transfer, FromEntry: fromEntry, ToEntry: toEntry},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker[*token.PasetoPayload]) {
				addAuthorizationHeader(t, accountFrom.Owner, request, authorizationHeaderKey, authorizationHeaderType, tokenMaker, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, param db.TransferTxParam, txRes db.TransferTxResult) {
				accountWithMismatchedCurrency := accountFrom
				for {
					curr := util.RandomCurrency()
					if curr != currency {
						accountWithMismatchedCurrency.Currency = curr
						break
					}
				}
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(accountFrom.ID)).Times(1).Return(accountWithMismatchedCurrency, nil)
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
			tc.buildStubs(store, tc.txParam, tc.txRes)

			// start test server request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			payloadJson := new(bytes.Buffer)

			util.Serialize(tc.req, payloadJson)

			request, err := http.NewRequest(http.MethodPost, "/transfer", payloadJson)
			require.NoError(t, err)
			// add authorization header
			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})

	}

}
