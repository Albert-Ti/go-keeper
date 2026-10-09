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

func (g *GrpcServer) DeleteArbitraryData(ctx context.Context, in *pb.DeleteArbitraryDataRequest) (*pb.DeleteArbitraryDataResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.Svc.DeleteArbitraryData(ctx, uuid, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "arbitrary data not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete arbitrary data")
	}

	return &pb.DeleteArbitraryDataResponse{}, nil
}
