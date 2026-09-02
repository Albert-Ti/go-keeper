package handler

import (
	"context"

	mytoken "github.com/Albert-Ti/go-keeper/internal/token"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := g.Svc.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		return nil, err
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
