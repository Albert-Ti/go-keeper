package handler

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-keeper/internal/service"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GrpcServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if len(in.GetPass()) < 3 {
		return nil, status.Errorf(codes.InvalidArgument, "pass: %v, is too short", in.GetPass())
	}

	err := g.Svc.Register(ctx, in.GetEmail(), in.GetPass())
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user: %s, is already registered", in.GetEmail())
		}
		return nil, status.Error(codes.Internal, "internal server")
	}

	response := pb.RegisterResponse_builder{}.Build()

	return response, nil
}
