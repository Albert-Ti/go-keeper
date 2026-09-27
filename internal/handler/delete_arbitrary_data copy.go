package handler

import (
	"context"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

func (g *GrpcServer) DeleteArbitraryData(ctx context.Context, in *pb.DeleteArbitraryDataRequest) (*pb.DeleteArbitraryDataResponse, error) {
	return &pb.DeleteArbitraryDataResponse{}, nil
}
