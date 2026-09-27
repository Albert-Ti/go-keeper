package handler

import (
	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"google.golang.org/grpc"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

type GrpcServer struct {
	pb.UnimplementedGoKeeperServiceServer
	Srv  *grpc.Server
	Svc  *service.Service
	Opts *config.Options
}
