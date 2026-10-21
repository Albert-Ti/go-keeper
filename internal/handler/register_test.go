package handler_test

import (
	"context"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	tests := []struct {
		name      string
		wantCode  codes.Code
		pass      string
		setupMock func(mock *mocks.MockUserRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			pass:     "12345",
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().
					AddUser(gomock.Any(), "example@mail.com", gomock.Any(), gomock.Any()).
					Return(nil)

			},
		},

		{
			name:     "user already exist",
			wantCode: codes.AlreadyExists,
			pass:     "12345",
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().
					AddUser(gomock.Any(), "example@mail.com", gomock.Any(), gomock.Any()).
					Return(service.ErrAlreadyExists)
			},
		},

		{
			name:     "pass is too short",
			wantCode: codes.InvalidArgument,
			pass:     "12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			req := pb.RegisterRequest_builder{Email: "example@mail.com", Pass: tt.pass}.Build()
			_, err := client.Register(context.Background(), req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
