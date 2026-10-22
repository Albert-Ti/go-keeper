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

func TestCreateCard(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCardRepo(ctrl)
	opts := config.NewOptions()
	cardSvc := service.NewCardService(mockRepo)

	client := NewTestGRPCServer(t, opts, nil, cardSvc, nil)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name       string
		cardNumber string
		expireDate string
		ctx        context.Context
		wantCode   codes.Code
		setupMock  func(mock *mocks.MockCardRepo)
	}{
		{
			name:       "success",
			cardNumber: "2200000000000053",
			expireDate: "01/27",
			ctx:        metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			wantCode:   codes.OK,
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().CreateCard(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:       "success another card",
			cardNumber: "6011111111111117",
			expireDate: "01/27",
			ctx:        metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			wantCode:   codes.OK,
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().CreateCard(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:       "invalid card number, IsLuhnValid ",
			cardNumber: "5555640328235481",
			expireDate: "01/27",
			ctx:        metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			wantCode:   codes.InvalidArgument,
		},
		{
			name:       "invalid card number, must be 16 digits",
			cardNumber: "55556403",
			expireDate: "01/27",
			ctx:        metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			wantCode:   codes.InvalidArgument,
		},
		{
			name:       "invalid expiry date format, expected MM/YY",
			cardNumber: "5555640328235487",
			expireDate: "01127",
			wantCode:   codes.InvalidArgument,
			ctx:        metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
		},
		{
			name:       "card expiry date has already passed",
			cardNumber: "5555640328235487",
			expireDate: "01/20",
			wantCode:   codes.InvalidArgument,
			ctx:        metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
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

			_, err := client.CreateCard(tt.ctx, pb.CreateCardRequest_builder{
				CardNumber: []byte(tt.cardNumber),
				ExpiryDate: tt.expireDate,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}

func TestGetCards(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCardRepo(ctrl)
	opts := config.NewOptions()
	cardSvc := service.NewCardService(mockRepo)

	client := NewTestGRPCServer(t, opts, nil, cardSvc, nil)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(mock *mocks.MockCardRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().GetCards(gomock.Any(), gomock.Any()).
					Return([]models.Card{}, nil)
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

			_, err := client.GetCards(tt.ctx, pb.CardsRequest_builder{}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}

func TestActivateCard(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCardRepo(ctrl)
	opts := config.NewOptions()
	cardSvc := service.NewCardService(mockRepo)

	client := NewTestGRPCServer(t, opts, nil, cardSvc, nil)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(mock *mocks.MockCardRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().ActivateCard(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:     "not found",
			wantCode: codes.NotFound,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().ActivateCard(gomock.Any(), gomock.Any(), gomock.Any()).
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
				tt.setupMock(mockRepo)
			}

			_, err := client.ActivateCard(tt.ctx, pb.ActiveCardRequest_builder{
				Id: 1,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}

func TestDeleteCard(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCardRepo(ctrl)
	opts := config.NewOptions()
	cardSvc := service.NewCardService(mockRepo)

	client := NewTestGRPCServer(t, opts, nil, cardSvc, nil)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(mock *mocks.MockCardRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().DeleteCard(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name:     "not found",
			wantCode: codes.NotFound,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockCardRepo) {
				mock.EXPECT().DeleteCard(gomock.Any(), gomock.Any(), gomock.Any()).
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
				tt.setupMock(mockRepo)
			}

			_, err := client.DeleteCard(tt.ctx, pb.DeleteCardRequest_builder{
				Id: 1,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}
