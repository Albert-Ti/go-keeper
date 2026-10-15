package handler

import (
	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) ReloadArbitraryData(stream pb.GoKeeperService_ReloadArbitraryDataServer) error {
	ctx := stream.Context()
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "failed to get user")
	}

	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "error while reading the stream from the server: %v", err)
	}

	reader := &utils.GoKeeperStream[*pb.ReloadArbitraryDataRequest]{Stream: stream}

	if err := g.Svc.ReloadArbitraryData(ctx, uuid, req.GetId(), reader); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(pb.ReloadArbitraryDataResponse_builder{}.Build())
}
