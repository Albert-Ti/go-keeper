package handler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/models"
	"github.com/Albert-Ti/go-keeper/internal/repository/mocks"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/token"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	tests := []struct {
		name      string
		wantCode  codes.Code
		pass      string
		setupMock func(mock *mocks.MockUserRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			pass:     "12345",
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().
					AddUser(gomock.Any(), "example@mail.com", gomock.Any(), gomock.Any()).
					Return(nil)

			},
		},

		{
			name:     "user already exist",
			wantCode: codes.AlreadyExists,
			pass:     "12345",
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().
					AddUser(gomock.Any(), "example@mail.com", gomock.Any(), gomock.Any()).
					Return(service.ErrAlreadyExists)
			},
		},

		{
			name:     "pass is too short",
			wantCode: codes.InvalidArgument,
			pass:     "12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			req := pb.RegisterRequest_builder{Email: "example@mail.com", Pass: tt.pass}.Build()
			_, err := client.Register(context.Background(), req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	testEmail := "example@mail.com"
	testPass := "12345"
	validHashPass := utils.HashString("salt", testPass)
	tests := []struct {
		name      string
		wantCode  codes.Code
		setupMock func(mock *mocks.MockUserRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{
						IsConfirmEmail: true,
						Pass:           validHashPass,
					}, nil)

			},
		},

		{
			name:     "not found",
			wantCode: codes.NotFound,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{}, service.ErrNoRows)

			},
		},

		{
			name:     "email invalid email or pass",
			wantCode: codes.Unauthenticated,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{}, service.ErrInvalidPassword)

			},
		},

		{
			name:     "email has not been confirmed",
			wantCode: codes.PermissionDenied,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetUserByEmail(gomock.Any(), testEmail).
					Return(models.User{
						IsConfirmEmail: false,
					}, service.ErrEmailNotConfirmed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockRepo)

			req := pb.LoginRequest_builder{Email: testEmail, Pass: testPass}.Build()
			_, err := client.Login(context.Background(), req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}
}

func TestRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	tokenCurr, err := token.CreateRefreshToken("user-1", opts.JWTSecret, opts.ExpiresRToken)
	require.NoError(t, err)

	tests := []struct {
		name         string
		refreshToken string
		wantCode     codes.Code
		setupMock    func(mock *mocks.MockUserRepo)
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

func TestConfirmEmail(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	tests := []struct {
		name      string
		wantCode  codes.Code
		email     string
		codeEmail string
		setupMock func(mock *mocks.MockUserRepo)
	}{
		{
			name:      "success",
			email:     "example@mail.com",
			codeEmail: "code123",
			wantCode:  codes.OK,
			setupMock: func(mock *mocks.MockUserRepo) {
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
			setupMock: func(mock *mocks.MockUserRepo) {
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

func TestGetProfile(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		ctx       context.Context
		setupMock func(mock *mocks.MockUserRepo)
	}{
		{
			name:     "success",
			wantCode: codes.OK,
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{}, nil)
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

			_, err := client.GetProfile(tt.ctx, pb.ProfileRequest_builder{}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}

func TestChangePass(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepo(ctrl)
	opts := config.NewOptions()
	userSvc := service.NewUserService(mockRepo, nil)

	client := NewTestGRPCServer(t, opts, userSvc, nil, nil)

	accessToken, err := token.CreateAccessToken("user-1", opts.JWTSecret, opts.ExpiresAToken)
	require.NoError(t, err)

	salt := "salt"
	hashPassCurr := utils.HashString(salt, "12345")
	hashPassNew := utils.HashString(salt, "54321")
	hashPassOld := utils.HashString(salt, "qwerty")

	tests := []struct {
		name      string
		wantCode  codes.Code
		passCurr  string
		passNew   string
		ctx       context.Context
		setupMock func(mock *mocks.MockUserRepo)
	}{
		{
			name:     "success",
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			passCurr: "12345",
			passNew:  "54321",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{Pass: hashPassCurr}, nil)

				mock.EXPECT().GetOldPassList(gomock.Any(), gomock.Any()).
					Return([]string{hashPassOld}, nil)

				mock.EXPECT().ChangePass(gomock.Any(), gomock.Any(), hashPassCurr, hashPassNew).
					Return(nil)
			},
		},
		{
			name:     "current password is incorrect",
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			passCurr: "1234",
			passNew:  "54321",
			wantCode: codes.FailedPrecondition,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{Pass: hashPassCurr}, nil)
			},
		},
		{
			name:     "password was used before",
			ctx:      metadata.AppendToOutgoingContext(context.Background(), "authorization", accessToken),
			passCurr: "12345",
			passNew:  "qwerty",
			wantCode: codes.FailedPrecondition,
			setupMock: func(mock *mocks.MockUserRepo) {
				mock.EXPECT().GetProfile(gomock.Any(), gomock.Any()).
					Return(models.Profile{Pass: hashPassCurr}, nil)

				mock.EXPECT().GetOldPassList(gomock.Any(), gomock.Any()).
					Return([]string{hashPassOld}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			_, err := client.ChangePass(tt.ctx, pb.PassRequest_builder{
				PassCurr: tt.passCurr,
				PassNew:  tt.passNew,
			}.Build())

			st, ok := status.FromError(err)
			require.True(t, ok)
			fmt.Println(st.Code().String())
			require.Equal(t, tt.wantCode, st.Code())
		})
	}

}
