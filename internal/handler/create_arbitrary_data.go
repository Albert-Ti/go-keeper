package handler

import (
	"fmt"
	"io"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

func (g *GrpcServer) CreateArbitraryData(stream pb.GoKeeperService_CreateArbitraryDataServer) error {

	var metadata *pb.FileMetadata
	var chunks []byte

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка при чтении стрима от сервера: %v", err)
		}

		if req.GetMetadata() != nil {
			metadata = req.GetMetadata()
			fmt.Println("metadata:", metadata)
		}

		if req.GetChunk() != nil {
			chunks = append(chunks, req.GetChunk()...)
			fmt.Println("chunk:", len(req.GetChunk()))
		}
	}
	fmt.Println("stream off")

	return nil
}
