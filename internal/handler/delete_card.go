package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/service"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) DeleteCard(ctx context.Context, in *pb.DeleteCardRequest) (*pb.DeleteCardResponse, error) {

	err := g.Svc.DeleteCard(ctx, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrCardNotFound) {
			return nil, status.Error(codes.NotFound, "card not found or not active")
		}
		return nil, status.Error(codes.Internal, "failed to delete card")
	}
	return &pb.DeleteCardResponse{}, nil
}
