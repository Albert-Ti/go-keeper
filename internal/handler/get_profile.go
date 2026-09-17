package handler

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *GrpcServer) GetProfile(ctx context.Context, in *pb.ProfileRequest) (*pb.ProfileResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")

	}
	profile, err := g.Svc.GetProfile(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server")
	}

	response := pb.ProfileResponse_builder{
		Email:     profile.Email,
		CreatedAt: timestamppb.New(profile.CreatedAt),
	}.Build()

	return response, nil
}
