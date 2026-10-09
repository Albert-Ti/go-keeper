package handler

import (
	"context"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *GrpcServer) GetArbitraryData(ctx context.Context, in *pb.ListArbitraryDataRequest) (*pb.ListArbitraryDataResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	arbitraryData, err := g.Svc.GetArbitraryData(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get cards")
	}

	var list []*pb.ArbitraryData
	for _, v := range arbitraryData {
		list = append(list, pb.ArbitraryData_builder{
			Id:         v.ID,
			Name:       v.Name,
			Type:       v.Type,
			Status:     v.Status,
			ObjectKey:  v.ObjectKey,
			OsPath:     v.OSPath,
			ClientSize: v.ClientSize,
			TotalSize:  v.TotalSize,
			CreatedAt:  timestamppb.New(v.CreatedAt),
		}.Build())
	}

	response := pb.ListArbitraryDataResponse_builder{
		ArbitraryData: list,
	}.Build()

	return response, nil
}
