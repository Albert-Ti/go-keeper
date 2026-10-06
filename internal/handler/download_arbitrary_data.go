package handler

import (
	"io"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) DownloadArbitraryData(in *pb.DownloadArbitraryDataRequest, stream pb.GoKeeperService_DownloadArbitraryDataServer) error {
	ctx := stream.Context()
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "failed to get user")
	}

	buf := make([]byte, 1024*1024)
	reader, err := g.Svc.DownloadArbitraryData(ctx, uuid, in.GetId())
	if err != nil {
		return status.Error(codes.Internal, "failed to download file")
	}

	defer reader.Close()

	for {
		n, err := reader.Read(buf)

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return status.Error(codes.Internal, "failed to download file")
		}
		if n > 0 {
			err := stream.Send(pb.DownloadArbitraryDataResponse_builder{Chunk: buf[:n]}.Build())

			if err != nil {
				return status.Error(codes.Internal, "failed to download file")
			}
		}
	}
}
