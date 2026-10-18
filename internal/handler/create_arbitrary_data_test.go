package handler_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
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
