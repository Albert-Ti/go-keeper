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

	db, err := repository.NewDatabase(opts.DBConnStr)
	if err != nil {
		panic(err)
	}
	cache := repository.NewCache(opts.CacheClientRunAddr, opts.CacheClientPass)
	cache.Ping(context.Background())

	cachedDB := repository.NewCachedDatabase(db, cache)

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

	objStorage, err := repository.NewObjStorage(
		opts.ObjStorageRunAddr, opts.ObjStorageAccessKey, opts.ObjStorageSecretKey)
	if err != nil {
		panic(err)
	}

	svc := service.NewService(cachedDB, opts, sender, objStorage)

	lis, err := net.Listen("tcp", opts.RunAddr)

	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // под 1MB чанки
		grpc.ChainUnaryInterceptor(
			interceptor.Logging(), interceptor.AuthUnary(opts.JWTSecret)),
		grpc.ChainStreamInterceptor(
			interceptor.AuthStream(opts.JWTSecret),
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
