package handler

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) CreateArbitraryData(ctx context.Context, in *pb.CreateArbitraryDataRequest) (*pb.CreateArbitraryDataResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	if err := g.Svc.CreateArbitraryData(ctx, uuid, in.GetFilename()); err != nil {
		return nil, status.Error(codes.Internal, "failed to create arbitrary data")
	}

	return &pb.CreateArbitraryDataResponse{}, nil
}
