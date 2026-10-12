package handler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/token"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestChangePass(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDatabase(ctrl)
	opts := config.NewOptions()
	svc := service.NewService(mockRepo, opts, nil, nil)

	client := NewTestGRPCServer(t, svc, opts)

	token, err := token.CreateAccessToken("user-1", opts.JWTSecret)
	require.NoError(t, err)

	salt := "salt"
	hashPassCurr := utils.HashString(salt, "12345")
	hashPassNew := utils.HashString(salt, "54321")
	hashPassOld := utils.HashString(salt, "qwerty")

	tests := []struct {
		name      string
		wantCode  codes.Code
		passCurr  string
		passNew   string
		ctx       context.Context
		setupMock func(mock *mocks.MockDatabase)
	}{
		{
			name:     "success",
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", token),
			passCurr: "12345",
			passNew:  "54321",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{Pass: hashPassCurr}, nil)

				mock.EXPECT().GetOldPassList(gomock.Any(), gomock.Any()).
					Return([]string{hashPassOld}, nil)

				mock.EXPECT().ChangePass(gomock.Any(), gomock.Any(), hashPassCurr, hashPassNew).
					Return(nil)
			},
		},
		{
			name:     "current password is incorrect",
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", token),
			passCurr: "1234",
			passNew:  "54321",
			wantCode: codes.FailedPrecondition,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{Pass: hashPassCurr}, nil)
			},
		},
		{
			name:     "password was used before",
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", token),
			passCurr: "12345",
			passNew:  "qwerty",
			wantCode: codes.FailedPrecondition,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{Pass: hashPassCurr}, nil)

				mock.EXPECT().GetOldPassList(gomock.Any(), gomock.Any()).
					Return([]string{hashPassOld}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			_, err := client.ChangePass(tt.ctx, pb.PassRequest_builder{
				PassCurr: tt.passCurr,
				PassNew:  tt.passNew,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			fmt.Println(st.Code().String())
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}
