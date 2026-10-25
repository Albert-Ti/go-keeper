package handler_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
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

func TestCreateArbitraryData(t *testing.T) {
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
				mDB.EXPECT().CreateArbitraryData(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
					Return(int64(1), nil)

				mObj.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mDB.EXPECT().UpdateArbitraryData(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
					Return(nil)
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

			stream, err := client.CreateArbitraryData(tt.ctx)
			require.NoError(t, err)

			// 31238 kb
			f, err := os.Open("test_data/voina-i-mir.pdf")
			require.NoError(t, err)
			defer f.Close()

			info, err := f.Stat()
			require.NoError(t, err)

			errSend := stream.Send(pb.CreateArbitraryDataRequest_builder{
				Metadata: pb.FileMetadata_builder{
					Filename: []byte("voina-i-mir"),
					Type:     "pdf",
					Size:     info.Size(),
					OsPath:   []byte("testdata/voina-i-mir.pdf"),
				}.Build(),
			}.Build())
			if errSend != io.EOF {
				require.NoError(t, errSend)
			}

			buf := make([]byte, 1024*1024)

			for {
				n, readErr := f.Read(buf)
				if !errors.Is(readErr, io.EOF) {
					require.NoError(t, readErr)
				}

				if n > 0 {
					err := stream.Send(pb.CreateArbitraryDataRequest_builder{
						Chunk: buf[:n],
					}.Build())
					if !errors.Is(err, io.EOF) {
						require.NoError(t, err)
					}
				}

				if readErr == io.EOF {
					break
				}
			}

			_, err = stream.CloseAndRecv()
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}

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

func TestReloadArbitraryData(t *testing.T) {
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

				mObj.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				mDB.EXPECT().UpdateArbitraryData(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).
					Return(nil)
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

			stream, err := client.ReloadArbitraryData(tt.ctx)
			require.NoError(t, err)

			// 31238 kb
			f, err := os.Open("test_data/voina-i-mir.pdf")
			require.NoError(t, err)
			defer f.Close()

			buf := make([]byte, 1024*1024)

			for {
				n, readErr := f.Read(buf)
				if !errors.Is(readErr, io.EOF) {
					require.NoError(t, readErr)
				}

				if n > 0 {
					err := stream.Send(pb.ReloadArbitraryDataRequest_builder{
						Chunk: buf[:n],
					}.Build())
					if !errors.Is(err, io.EOF) {
						require.NoError(t, err)
					}
				}

				if readErr == io.EOF {
					break
				}
			}

			_, err = stream.CloseAndRecv()
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}

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

func TestDeleteArbitraryData(t *testing.T) {

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
			setupMock: func(mDB *mocks.MockDataRepo, mObj *mocks.MockObjStorage) {
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
