package handler

import (
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

func (g *GrpcServer) ReloadArbitraryData(stream pb.GoKeeperService_ReloadArbitraryDataServer) error {
	return stream.SendAndClose(pb.ReloadArbitraryDataResponse_builder{}.Build())
}
