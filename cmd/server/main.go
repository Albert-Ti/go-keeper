package main

import (
	"context"
	"log/slog"
	"net"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/email"
	"github.com/Albert-Ti/go-keeper/internal/handler"
	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/repository"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"google.golang.org/grpc"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

func main() {
	opts, errCfg := config.Build()
	if errCfg != nil {
		panic(errCfg)
	}

	pgRepo, err := repository.NewDatabase(opts.DBConnStr)
	if err != nil {
		panic(err)
	}
	cache := repository.NewCache("localhost:6379", "redis")
	cache.Ping(context.Background())

	repo := repository.NewCachedDatabase(pgRepo, cache)

	var sender *email.Sender
	if opts.EnableSMTP {
		sender, err = email.NewSender(
			opts.SMTPOpt.Host,
			opts.SMTPOpt.Port,
			opts.SMTPOpt.Username,
			opts.SMTPOpt.Pass,
			opts.SMTPOpt.Username,
		)
		if err != nil {
			panic(err)
		}
	}

	objectStorage, err := repository.NewObjectStorage("http://localhost:8333", "gokeeper-access", "gokeeper-secret")
	if err != nil {
		panic(err)
	}

	svc := service.NewService(repo, opts, sender, objectStorage)

	lis, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.Logging(), interceptor.Auth(opts.JWTSecret),
		),
	)

	pb.RegisterGoKeeperServiceServer(srv, &handler.GrpcServer{
		Svc:  svc,
		Opts: opts,
	})

	slog.Info("running server", "host", opts.RunAddr)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}

}
