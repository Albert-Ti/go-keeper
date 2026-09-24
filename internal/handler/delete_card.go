package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/service"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) DeleteCard(ctx context.Context, in *pb.DeleteCardRequest) (*pb.DeleteCardResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.Svc.DeleteCard(ctx, uuid, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrCardNotFound) {
			return nil, status.Error(codes.NotFound, "card not found or not active")
		}
		return nil, status.Error(codes.Internal, "failed to delete card")
	}
	return &pb.DeleteCardResponse{}, nil
}
