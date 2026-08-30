package handler

import (
	"context"
	"fmt"
	"strconv"

	mytoken "github.com/Albert-Ti/go-keeper/internal/token"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (g GrpcServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, err := g.Svc.Register(ctx, in.GetEmail(), in.GetPassword())
	fmt.Println(err)
	if err != nil {
		return nil, err
	}

	token, err := mytoken.CreateToken(strconv.Itoa(userID), g.Opts.JWTSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create token")
	}

	err = grpc.SetHeader(ctx, metadata.Pairs("authorization", token))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set header")
	}

	return &pb.RegisterResponse{}, nil
}
