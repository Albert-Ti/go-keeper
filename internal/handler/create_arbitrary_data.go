package handler

import (
	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) CreateArbitraryData(stream pb.GoKeeperService_CreateArbitraryDataServer) error {
	var metadata *pb.FileMetadata
	ctx := stream.Context()
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "failed to get user")
	}

	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "error while reading the stream from the server: %v", err)
	}

	if req.GetMetadata() != nil {
		metadata = req.GetMetadata()
	}

	reader := &utils.GoKeeperStream[*pb.CreateArbitraryDataRequest]{
		Stream: stream,
	}

	id, err := g.Svc.SaveArbitraryData(
		ctx,
		uuid,
		metadata.GetFilename(),
		metadata.GetType(),
		metadata.GetOsPath(),
		metadata.GetSize(),
		reader,
	)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(pb.CreateArbitraryDataResponse_builder{
		Id: id,
	}.Build())
}
