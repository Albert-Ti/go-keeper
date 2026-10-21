package handler_test

import (
	"bytes"
	"context"
	"io"
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

func TestDownloadArbitraryData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDataRepo(ctrl)
	mockObj := mocks.NewMockObjStorage(ctrl)
	opts := config.NewOptions()
	dataSvc := service.NewDataService(mockRepo, mockObj)

	client := NewTestGRPCServer(t, opts, nil, nil, dataSvc)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(*mocks.MockDataRepo, *mocks.MockObjStorage)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mDB *mocks.MockDataRepo, mObj *mocks.MockObjStorage) {
				mDB.EXPECT().GetArbitraryDataByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(models.ArbitraryData{}, nil)

				content := []byte("test data")
				mObj.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(io.NopCloser(bytes.NewReader(content)), nil)
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

			stream, err := client.DownloadArbitraryData(tt.ctx, pb.DownloadArbitraryDataRequest_builder{
				Id: 1,
			}.Build())

			// 1. Проверяем ошибку открытия стрима
			if err != nil {
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantCode, st.Code())
				return
			}

			// 2. Читаем стрим до конца
			var recvErr error
			for {
				_, recvErr = stream.Recv()
				if recvErr != nil {
					break
				}
			}

			// 3. io.EOF = нормальное завершение = codes.OK
			if recvErr == io.EOF {
				require.Equal(t, codes.OK, tt.wantCode)
				return
			}

			// 4. Иначе — реальная ошибка
			st, ok := status.FromError(recvErr)
			require.True(t, ok, "expected grpc status, got: %v", recvErr)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
