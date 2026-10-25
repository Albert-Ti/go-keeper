package handler

import (
	"context"
	"errors"
	"io"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/Albert-Ti/go-keeper/internal/utils"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	id, err := g.DataSvc.SaveArbitraryData(
		ctx,
		uuid,
		metadata.GetFilename(),
		metadata.GetOsPath(),
		metadata.GetType(),
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

func (g *GrpcServer) GetArbitraryData(ctx context.Context, in *pb.ListArbitraryDataRequest) (*pb.ListArbitraryDataResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}
	arbitraryData, err := g.DataSvc.GetArbitraryData(ctx, uuid)
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

	if err := g.DataSvc.ReloadArbitraryData(ctx, uuid, req.GetId(), reader); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(pb.ReloadArbitraryDataResponse_builder{}.Build())
}

func (g *GrpcServer) DownloadArbitraryData(in *pb.DownloadArbitraryDataRequest, stream pb.GoKeeperService_DownloadArbitraryDataServer) error {
	ctx := stream.Context()
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "failed to get user")
	}

	buf := make([]byte, 1024*1024)
	reader, totalSize, err := g.DataSvc.DownloadArbitraryData(ctx, uuid, in.GetId())
	if err != nil {
		return status.Error(codes.Internal, "failed to download file")
	}

	defer reader.Close()

	errSend := stream.Send(pb.DownloadArbitraryDataResponse_builder{
		Metadata: pb.FileMetadata_builder{
			Size: totalSize,
		}.Build(),
	}.Build())
	if errSend != nil {
		return status.Error(codes.Internal, "failed to download file")
	}

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

func (g *GrpcServer) DeleteArbitraryData(ctx context.Context, in *pb.DeleteArbitraryDataRequest) (*pb.DeleteArbitraryDataResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	err = g.DataSvc.DeleteArbitraryData(ctx, uuid, in.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "arbitrary data not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete arbitrary data")
	}

	return &pb.DeleteArbitraryDataResponse{}, nil
}
