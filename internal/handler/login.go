package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/service"
	mytoken "github.com/Albert-Ti/go-keeper/internal/token"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := g.Svc.Login(ctx, in.GetEmail(), in.GetPass())
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
		return nil, status.Error(codes.Internal, "internal server")
	}
	accessToken, err := mytoken.CreateAccessToken(user.UUID, g.Opts.JWTSecret)
	refreshToken, err := mytoken.CreateRefreshToken(user.UUID, g.Opts.JWTSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create token")
	}
	response := pb.LoginResponse_builder{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}.Build()

	return response, nil
}
