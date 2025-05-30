package api

import (
	"bytes"
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

func GetRandomUser() db.User {
	return db.User{
		Username:       util.RandomString(6),
		FullName:       util.RandomString(16),
		Email:          util.RandomEmail(),
		HashedPassword: util.RandomString(16),
	}
}

func TestCreateUserApi(t *testing.T) {
	user := GetRandomUser()
	password := util.RandomString(6)
	hashedPass, _ := util.HashPassword(password)
	fmt.Println("hashedPass:", hashedPass)
	user.HashedPassword = hashedPass
	fmt.Println("user.HashedPassword:", user.HashedPassword)

	testCases := []struct {
		name       string
		req        createUserRequest
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			req:  createUserRequest{Username: user.Username, FullName: user.FullName, Email: user.Email, Password: password},
			buildStubs: func(store *mockdb.MockStore) {
				// param := db.CreateUserParams{Username: user.Username, FullName: user.FullName, Email: user.Email, HashedPassword: hashedPass}
				store.EXPECT().CreateUser(gomock.Any(), gomock.All()).Times(1).Return(user, nil)
			},
			checkResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var respUser db.User
				fmt.Println("recorder.Body.Bytes()", string(recorder.Body.Bytes()))
				require.Equal(t, http.StatusOK, recorder.Code)
				err := json.Unmarshal(recorder.Body.Bytes(), &respUser)
				require.NoError(t, err)
				require.EqualValues(t, user, respUser)
			},
		}}

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

			payloadJson := new(bytes.Buffer)
			util.Serialize(tc.req, payloadJson)
			fmt.Println("payloadJson:", payloadJson)

			request, err := http.NewRequest(http.MethodPost, "/users", payloadJson)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})

	}

}

func TestGetUserApi(t *testing.T) {
	user := GetRandomUser()
	testCases := []struct {
		name       string
		username   string
		buildStubs func(store *mockdb.MockStore)
		checkResp  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			username: user.Username,
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

			url := fmt.Sprintf("/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			// check response
			tc.checkResp(t, recorder)

		})

	}

}
