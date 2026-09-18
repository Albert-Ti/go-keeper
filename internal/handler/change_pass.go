package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/service"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) ChangePass(ctx context.Context, in *pb.PassRequest) (*pb.PassResponse, error) {
	uuid, err := interceptor.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user")
	}

	if err := g.Svc.ChangePass(ctx, uuid, in.GetPassOld(), in.GetPassNew()); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			return nil, status.Error(codes.FailedPrecondition, "old password is incorrect")
		}
		return nil, status.Error(codes.Internal, "failed to change password")
	}

	return &pb.PassResponse{}, nil
}
