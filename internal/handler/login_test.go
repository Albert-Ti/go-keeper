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

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	svc := service.NewService(mockRepo)

	opts := config.NewOptions()
	client := NewTestGRPCServer(t, svc, opts)

	tests := []struct {
		name      string
		wantCode  codes.Code
		setupMock func(mock *mocks.MockRepository)
	}{
		{
			name:     "Success",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().GetUser(gomock.Any(), "example@mail.com").
					Return(models.User{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockRepo)

			req := pb.LoginRequest_builder{Email: "example@mail.com", Password: "12345"}.Build()
			_, err := client.Login(context.Background(), req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
