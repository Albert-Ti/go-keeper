package handler

import (
	"fmt"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StreamReader struct {
	stream pb.GoKeeperService_CreateArbitraryDataServer
	buf    []byte
	Total  int64
}

func (r *StreamReader) Read(p []byte) (int, error) {

	for len(r.buf) == 0 {
		req, err := r.stream.Recv()
		fmt.Println("CHUNK", len(req.GetChunk()))
		if err != nil {
			return 0, err
		}

		if chunk := req.GetChunk(); len(chunk) > 0 {
			r.buf = chunk
		}
	}

	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	r.Total += int64(n)

	return n, nil
}

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

	reader := &StreamReader{stream: stream}

	id, err := g.Svc.SaveArbitraryData(
		ctx,
		uuid,
		metadata.GetFilename(),
		metadata.GetType(),
		reader,
		metadata.GetSize(),
		reader.Total,
	)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(pb.CreateArbitraryDataResponse_builder{
		Id: id,
	}.Build())
}
