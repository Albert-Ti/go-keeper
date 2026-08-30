package handler

import (
	"context"
	"fmt"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

func (g GrpcServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := g.Svc.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		return nil, err
	}
	fmt.Println(user)

	return &pb.LoginResponse{}, nil
}
