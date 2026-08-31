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
	token, err := g.Svc.Register(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user %s is already registered", in.GetEmail())
		}
		return nil, status.Error(codes.Internal, "db error")
	}

	response := pb.RegisterResponse_builder{
		ConfirmToken: token,
	}.Build()

	return response, nil
}
