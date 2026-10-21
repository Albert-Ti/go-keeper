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
	Opts *config.Options

	UserSvc *service.UserService
	CardSvc *service.CardService
	DataSvc *service.DataService
}
