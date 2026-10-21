package handler_test

import (
	"context"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/token"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGetArbitraryData(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDataRepo(ctrl)
	opts := config.NewOptions()
	dataSvc := service.NewDataService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, nil, nil, dataSvc)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(mock *mocks.MockDataRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockDataRepo) {
				mock.EXPECT().GetArbitraryData(gomock.Any(), gomock.Any()).
					Return([]models.ArbitraryData{}, nil)
			},
		},

		{
			name:     "unauthenticated",
			wantCode: codes.Unauthenticated,
			ctx:      context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			_, err := client.GetArbitraryData(tt.ctx, pb.ListArbitraryDataRequest_builder{}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}
