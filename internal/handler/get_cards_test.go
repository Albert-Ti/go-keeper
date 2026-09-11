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

func TestGetCards(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	opts := config.NewOptions()

	svc := service.NewService(mockRepo, opts, nil)

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
				mock.EXPECT().GetCards(gomock.Any(), gomock.Any()).
					Return([]models.Card{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockRepo)

			_, err := client.GetCards(context.Background(), &pb.CardsRequest{})

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}
