package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/service"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) ConfirmEmail(ctx context.Context, in *pb.ConfirmEmailRequest) (*pb.ConfirmEmailResponse, error) {
	err := g.Svc.ConfirmEmail(ctx, in.GetEmail(), in.GetEmailCode())

	if err != nil {
		if errors.Is(err, service.ErrInvalidCodeEmail) {
			return nil, status.Errorf(codes.PermissionDenied, "invalid confirmation code: %v", in.GetEmailCode())
		}
		return nil, status.Error(codes.Internal, "failed to confirm email")
	}

	return &pb.ConfirmEmailResponse{}, nil
}
