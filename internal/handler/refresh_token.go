package handler

import (
	"context"

	mytoken "github.com/Albert-Ti/go-keeper/internal/token"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) RefreshToken(ctx context.Context, in *pb.TokenRequest) (*pb.TokenResponse, error) {
	claims := &mytoken.MyCustomClaims{}
	token, err := jwt.ParseWithClaims(in.GetRefreshToken(), claims, func(t *jwt.Token) (any, error) {
		return []byte(g.Opts.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "refresh token is expired")
	}

	accessToken, err := mytoken.CreateAccessToken(claims.UserID, g.Opts.JWTSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create access token")
	}
	refreshToken, err := mytoken.CreateRefreshToken(claims.UserID, g.Opts.JWTSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create refresh token")
	}

	response := pb.TokenResponse_builder{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}.Build()

	return response, nil
}
