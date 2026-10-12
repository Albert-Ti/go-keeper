package handler_test

import (
	"context"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDatabase(ctrl)
	opts := config.NewOptions()
	svc := service.NewService(mockRepo, opts, nil, nil)

	client := NewTestGRPCServer(t, svc, opts)

	testEmail := "example@mail.com"
	testPass := "12345"
	validHashPass := utils.HashString("salt", testPass)
	tests := []struct {
		name      string
		wantCode  codes.Code
		setupMock func(mock *mocks.MockDatabase)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{
						IsConfirmEmail: true,
						Pass:           validHashPass,
					}, nil)

			},
		},

		{
			name:     "not found",
			wantCode: codes.NotFound,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{}, service.ErrNoRows)

			},
		},

		{
			name:     "email invalid email or pass",
			wantCode: codes.Unauthenticated,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{}, service.ErrInvalidPassword)

			},
		},

		{
			name:     "email has not been confirmed",
			wantCode: codes.PermissionDenied,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{
						IsConfirmEmail: false,
					}, service.ErrEmailNotConfirmed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockRepo)

			req := pb.LoginRequest_builder{Email: testEmail, Pass: testPass}.Build()
			_, err := client.Login(context.Background(), req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
