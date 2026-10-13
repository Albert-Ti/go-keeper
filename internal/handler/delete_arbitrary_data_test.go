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

func TestDeleteArbitraryData(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDatabase(ctrl)
	mockObj := mocks.NewMockObjStorage(ctrl)
	opts := config.NewOptions()
	svc := service.NewService(mockRepo, opts, nil, mockObj)

	client := NewTestGRPCServer(t, svc, opts)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(*mocks.MockDatabase, *mocks.MockObjStorage)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mDB *mocks.MockDatabase, mObj *mocks.MockObjStorage) {
				mDB.EXPECT().GetArbitraryDataByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(models.ArbitraryData{}, nil)

				mObj.EXPECT().Delete(gomock.Any(), gomock.Any()).
					Return(nil)

				mDB.EXPECT().DeleteArbitraryData(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:     "not found",
			wantCode: codes.NotFound,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mDB *mocks.MockDatabase, mObj *mocks.MockObjStorage) {
				mDB.EXPECT().GetArbitraryDataByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(models.ArbitraryData{}, nil)

				mObj.EXPECT().Delete(gomock.Any(), gomock.Any()).
					Return(nil)

				mDB.EXPECT().DeleteArbitraryData(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(service.ErrNotFound)
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
				tt.setupMock(mockRepo, mockObj)
			}

			_, err := client.DeleteArbitraryData(tt.ctx, pb.DeleteArbitraryDataRequest_builder{
				Id: 1,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}
