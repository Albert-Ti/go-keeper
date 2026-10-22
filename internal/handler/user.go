package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/service"
	mytoken "github.com/Albert-Ti/go-keeper/internal/token"
	"github.com/golang-jwt/jwt/v5"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *GrpcServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(in.GetPass()) < 3 {
		return nil, status.Errorf(codes.InvalidArgument, "pass: %v, is too short", in.GetPass())
	}

	err := g.UserSvc.Register(ctx, in.GetEmail(), in.GetPass())
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user: %s, is already registered", in.GetEmail())
		}
		return nil, status.Error(codes.Internal, "failed to register")
	}

	response := pb.RegisterResponse_builder{}.Build()

	return response, nil
}

func (g *GrpcServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := g.UserSvc.Login(ctx, in.GetEmail(), in.GetPass())
	if err != nil {
		if errors.Is(err, service.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user: %v, not found", in.GetEmail())
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			return nil, status.Errorf(codes.Unauthenticated, "email: %v, invalid email or pass", in.GetEmail())
		}
		if errors.Is(err, service.ErrEmailNotConfirmed) {
			return nil, status.Errorf(codes.PermissionDenied, "email: %v, has not been confirmed", in.GetEmail())
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}

	accessToken, err := mytoken.CreateAccessToken(user.UUID, g.Opts.JWTSecret, g.Opts.ExpiresAToken)
	refreshToken, err := mytoken.CreateRefreshToken(user.UUID, g.Opts.JWTSecret, g.Opts.ExpiresRToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	response := pb.LoginResponse_builder{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}.Build()

	return response, nil
}

func (g *GrpcServer) RefreshToken(ctx context.Context, in *pb.TokenRequest) (*pb.TokenResponse, error) {
	claims := &mytoken.MyCustomClaims{}
	token, err := jwt.ParseWithClaims(in.GetRefreshToken(), claims, func(t *jwt.Token) (any, error) {
		return []byte(g.Opts.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "refresh token is expired")
	}

	accessToken, err := mytoken.CreateAccessToken(claims.UserID, g.Opts.JWTSecret, g.Opts.ExpiresAToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create access token")
	}
	refreshToken, err := mytoken.CreateRefreshToken(claims.UserID, g.Opts.JWTSecret, g.Opts.ExpiresRToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create refresh token")
	}

	response := pb.TokenResponse_builder{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}.Build()

	return response, nil
}

func (g *GrpcServer) GetProfile(ctx context.Context, in *pb.ProfileRequest) (*pb.ProfileResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")

	}
	profile, err := g.UserSvc.GetProfile(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	response := pb.ProfileResponse_builder{
		Email:     profile.Email,
		CreatedAt: timestamppb.New(profile.CreatedAt),
	}.Build()

	return response, nil
}

func (g *GrpcServer) ConfirmEmail(ctx context.Context, in *pb.ConfirmEmailRequest) (*pb.ConfirmEmailResponse, error) {
	err := g.UserSvc.ConfirmEmail(ctx, in.GetEmail(), in.GetEmailCode())

	if err != nil {
		if errors.Is(err, service.ErrInvalidCodeEmail) {
			return nil, status.Errorf(codes.PermissionDenied, "invalid confirmation code: %v", in.GetEmailCode())
		}
		return nil, status.Error(codes.Internal, "failed to confirm email")
	}

	return &pb.ConfirmEmailResponse{}, nil
}

func (g *GrpcServer) ChangePass(ctx context.Context, in *pb.PassRequest) (*pb.PassResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	if err := g.UserSvc.ChangePass(ctx, uuid, in.GetPassCurr(), in.GetPassNew()); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			return nil, status.Error(codes.FailedPrecondition, "current password is incorrect")
		}
		if errors.Is(err, service.ErrPasswordReused) {
			return nil, status.Error(codes.FailedPrecondition, "password was used before")
		}
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &pb.PassResponse{}, nil
}
