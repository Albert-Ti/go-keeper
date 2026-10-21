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

func (g *GrpcServer) ActivateCard(ctx context.Context, in *pb.ActiveCardRequest) (*pb.ActiveCardResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.CardSvc.ActivateCard(ctx, uuid, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "card not found")
		}
		return nil, status.Error(codes.Internal, "failed to activate card")
	}
	return &pb.ActiveCardResponse{}, nil
}
