package handler_test

import (
	"context"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/token"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDatabase(ctrl)
	opts := config.NewOptions()
	svc := service.NewService(mockRepo, opts, nil, nil)

	client := NewTestGRPCServer(t, svc, opts)

	tokenCurr, err := token.CreateRefreshToken("user-1", opts.JWTSecret)
	require.NoError(t, err)

	tests := []struct {
		name         string
		refreshToken string
		wantCode     codes.Code
		setupMock    func(mock *mocks.MockDatabase)
	}{
		{
			name:         "success",
			refreshToken: tokenCurr,
			wantCode:     codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := client.RefreshToken(context.Background(),
				pb.TokenRequest_builder{RefreshToken: tt.refreshToken}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
