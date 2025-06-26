package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	mockdb "github.com/bank_go/db/mock"
	db "github.com/bank_go/db/sqlc"
	pb_sources "github.com/bank_go/pb"
	mockqt "github.com/bank_go/queues/mock"
	"github.com/bank_go/token"
	"github.com/bank_go/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserRPCApi(t *testing.T) {
	user := GetRandomUser(t)
	newFullName := util.RandomString(6)
	newEmail := util.RandomEmail()
	wrongEmail := util.RandomString(10)

	testCases := []struct {
		name         string
		req          *pb_sources.UpdateUserRequest
		buildContext func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context
		buildStubs   func(store *mockdb.MockStore)
		checkResp    func(t *testing.T, res *pb_sources.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req:  &pb_sources.UpdateUserRequest{Username: user.Username, FullName: newFullName, Email: newEmail},
			buildContext: func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				param := db.UpdateUserParams{
					Username: user.Username,
					Email:    sql.NullString{String: newEmail, Valid: true},
					FullName: sql.NullString{String: newFullName, Valid: true},
				}
				updatedUser := db.User{
					Username:          user.Username,
					Email:             newEmail,
					FullName:          newFullName,
					HashedPassword:    user.HashedPassword,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(param)).Times(1).Return(updatedUser, nil)
			},
			checkResp: func(t *testing.T, res *pb_sources.UpdateUserResponse, err error) {
				require.NoError(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.OK, st.Code())
				respUser := res.GetUser()
				require.Equal(t, user.Username, respUser.Username)
				require.Equal(t, newEmail, respUser.Email)
				require.Equal(t, newFullName, respUser.FullName)
			},
		},
		{
			name: "UserNotFound",
			req:  &pb_sources.UpdateUserRequest{Username: user.Username, FullName: newFullName, Email: newEmail},
			buildContext: func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, res *pb_sources.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
				assert.ErrorContains(t, err, fmt.Sprintf("user with username '%s' not found", user.Username))
			},
		},
		{
			name: "TokenExpired",
			req:  &pb_sources.UpdateUserRequest{Username: user.Username, FullName: newFullName, Email: newEmail},
			buildContext: func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0).Return(db.User{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, res *pb_sources.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidAuth",
			req:  &pb_sources.UpdateUserRequest{Username: user.Username, FullName: newFullName, Email: newEmail},
			buildContext: func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context {
				return context.Background()

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0).Return(db.User{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, res *pb_sources.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "UsernameNotExist",
			req:  &pb_sources.UpdateUserRequest{Username: user.Username, FullName: newFullName, Email: newEmail},
			buildContext: func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, res *pb_sources.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req:  &pb_sources.UpdateUserRequest{Username: user.Username, FullName: newFullName, Email: wrongEmail},
			buildContext: func(t *testing.T, tokenMaker token.Maker[*token.PasetoPayload]) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0).Return(db.User{}, sql.ErrNoRows)
			},
			checkResp: func(t *testing.T, res *pb_sources.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
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
			tc.buildStubs(store)

			// start test server
			server := NewTestServer(t, store, qtProvider)
			ctx := tc.buildContext(t, server.tokenMaker)
			// ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.UpdateUser(ctx, tc.req)
			// check response
			tc.checkResp(t, res, err)

		})

	}

}
