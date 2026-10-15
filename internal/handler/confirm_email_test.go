package handler_test

import (
	"context"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConfirmEmail(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDatabase(ctrl)
	opts := config.NewOptions()
	svc := service.NewService(mockRepo, opts, nil, nil)

	client := NewTestGRPCServer(t, svc, opts)

	tests := []struct {
		name      string
		wantCode  codes.Code
		email     string
		codeEmail string
		setupMock func(mock *mocks.MockDatabase)
	}{
		{
			name:      "success",
			email:     "example@mail.com",
			codeEmail: "code123",
			wantCode:  codes.OK,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(models.User{
						Email:     "example@mail.com",
						EmailCode: "code123",
					}, nil)

				mock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},

		{
			name:      "permission denied",
			email:     "example@mail.com",
			codeEmail: "code123",
			wantCode:  codes.PermissionDenied,
			setupMock: func(mock *mocks.MockDatabase) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(models.User{}, service.ErrInvalidCodeEmail)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			_, err := client.ConfirmEmail(context.Background(), pb.ConfirmEmailRequest_builder{
				Email:     tt.email,
				EmailCode: tt.codeEmail,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}
