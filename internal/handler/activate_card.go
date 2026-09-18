package handler

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) ActivateCard(ctx context.Context, in *pb.ActiveCardRequest) (*pb.ActiveCardResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.Svc.ActivateCard(ctx, uuid, in.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to activate card")
	}
	return &pb.ActiveCardResponse{}, nil
}
