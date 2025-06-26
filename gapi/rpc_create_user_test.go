package gapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	mockdb "github.com/bank_go/db/mock"
	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	"github.com/bank_go/queues"
	mockqt "github.com/bank_go/queues/mock"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userParamTxEqMatcher struct {
	arg      db.CreateUserTxParam
	password string
	user     db.User
}

func (expected userParamTxEqMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParam)
	if !ok {
		return false
	}

	err := util.CheckPasswordHash(expected.password, actualArg.CreateUserParams.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.CreateUserParams.HashedPassword = actualArg.CreateUserParams.HashedPassword

	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	err = actualArg.AfterCreate(expected.user)
	return err == nil
}

func (e userParamTxEqMatcher) String() string {
	return fmt.Sprintf("is equal to %v (%T)", e.arg, e.arg)
}

func eqCreateUserTxParam(arg db.CreateUserTxParam, password string, user db.User) gomock.Matcher {
	return userParamTxEqMatcher{arg, password, user}
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

func TestCreateUserRPCApi(t *testing.T) {
	user := GetRandomUser(t)
	password := util.RandomString(6)
	hashedPass, err := util.HashPassword(password)
	require.NoError(t, err)
	user.HashedPassword = hashedPass
	queueErr := errors.New("enqueue verify email task error")

	testCases := []struct {
		name         string
		req          *pb_sources.CreateUserRequest
		buildContext func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context
		buildStubs   func(store *mockdb.MockStore, qtProvider *mockqt.MockTaskProvider)
		checkResp    func(t *testing.T, res *pb_sources.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req:  &pb_sources.CreateUserRequest{Username: user.Username, FullName: user.FullName, Email: user.Email, Password: password},
			// buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
			// 	return newContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute, token.TokenTypeAccessToken)
			// },
			buildStubs: func(store *mockdb.MockStore, qtProvider *mockqt.MockTaskProvider) {
				createUserParams := db.CreateUserParams{Username: user.Username, FullName: user.FullName, Email: user.Email, HashedPassword: user.HashedPassword}
				txParam := db.CreateUserTxParam{CreateUserParams: createUserParams}
				txResult := db.CreateUserTxResult{User: user}
				store.EXPECT().CreateUserTx(gomock.Any(), eqCreateUserTxParam(txParam, password, user)).Times(1).Return(txResult, nil)

				p := &queues.EmailVerifyPayload{Username: user.Username, FullName: user.FullName, Email: user.Email}
				qtProvider.EXPECT().ProvideEmailVerifyTask(gomock.Any(), gomock.Eq(p), gomock.Any()).Times(1).Return(nil)
			},
			checkResp: func(t *testing.T, res *pb_sources.CreateUserResponse, err error) {
				require.NoError(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.OK, st.Code())
			},
		},
		{
			name: "InternalDbErr",
			req:  &pb_sources.CreateUserRequest{Username: user.Username, FullName: user.FullName, Email: user.Email, Password: password},
			// buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
			// 	return newContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute, token.TokenTypeAccessToken)
			// },
			buildStubs: func(store *mockdb.MockStore, qtProvider *mockqt.MockTaskProvider) {
				txResult := db.CreateUserTxResult{}
				store.EXPECT().CreateUserTx(gomock.Any(), gomock.Any()).Times(1).Return(txResult, sql.ErrConnDone)
			},
			checkResp: func(t *testing.T, res *pb_sources.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "ProvideEmailVerifyTaskError",
			req:  &pb_sources.CreateUserRequest{Username: user.Username, FullName: user.FullName, Email: user.Email, Password: password},
			// buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
			// 	return newContextWithBearerToken(t, tokenMaker, user.Username, user.Role, time.Minute, token.TokenTypeAccessToken)
			// },
			buildStubs: func(store *mockdb.MockStore, qtProvider *mockqt.MockTaskProvider) {
				txResult := db.CreateUserTxResult{User: user}
				store.EXPECT().CreateUserTx(gomock.Any(), gomock.Any()).Times(1).Return(txResult, queueErr)
				qtProvider.EXPECT().ProvideEmailVerifyTask(gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(queueErr)

			},
			checkResp: func(t *testing.T, res *pb_sources.CreateUserResponse, err error) {
				fmt.Println("err:=> ", err)
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			storeCtl := gomock.NewController(t)
			defer storeCtl.Finish()
			store := mockdb.NewMockStore(storeCtl)

			queueCtl := gomock.NewController(t)
			defer queueCtl.Finish()
			qtProvider := mockqt.NewMockTaskProvider(queueCtl)
			// test build stubs
			tc.buildStubs(store, qtProvider)

			// start test server
			server := NewTestServer(t, store, qtProvider)
			// ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.CreateUser(context.Background(), tc.req)
			// check response
			tc.checkResp(t, res, err)

		})

	}

}
