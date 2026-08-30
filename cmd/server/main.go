package main

import (
	"log/slog"
	"net"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/handler"
	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/service"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc"
)

func main() {
	opts := config.NewOptions()

	repo, err := repository.NewRepository(opts.DBConnStr)
	if err != nil {
		panic(err)
	}
	svc := service.NewService(repo)

	lis, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.Logging()),
	)

	pb.RegisterGoKeeperServiceServer(srv, handler.GrpcServer{
		Svc:  svc,
		Opts: opts,
	})

	slog.Info("running server", "host", opts.RunAddr)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
