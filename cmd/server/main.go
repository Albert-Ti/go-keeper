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
	cached "github.com/Albert-Ti/go-keeper/internal/repository/cashed"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"google.golang.org/grpc"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

func main() {
	opts, errCfg := config.Build()
	if errCfg != nil {
		panic(errCfg)
	}

	db, err := repository.NewPostgres(opts.DBConnStr)
	if err != nil {
		panic(err)
	}
	cache := repository.NewCache(opts.CacheClientRunAddr, opts.CacheClientPass)
	cache.Ping(context.Background())

	userRepo := cached.NewUserCachedRepo(db, cache)
	cardRepo := cached.NewCardCachedRepo(db, cache)
	dataRepo := cached.NewDataCachedRepo(db, cache)

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
		opts.ObjStorageRunAddr,
		opts.ObjStorageAccessKey,
		opts.ObjStorageSecretKey,
		opts.ObjStorageBucket,
	)
	if err != nil {
		panic(err)
	}

	userSvc := service.NewUserService(userRepo, sender)
	cardSvc := service.NewCardService(cardRepo)
	dataSvc := service.NewDataService(dataRepo, objStorage)

	lis, err := net.Listen("tcp", opts.RunAddr)

	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024),
		grpc.ChainUnaryInterceptor(
			interceptor.Logging(), interceptor.AuthUnary(opts.JWTSecret)),
		grpc.ChainStreamInterceptor(
			interceptor.LoggingStream(), interceptor.AuthStream(opts.JWTSecret),
		),
	)

	pb.RegisterGoKeeperServiceServer(srv, &handler.GrpcServer{
		Opts:    opts,
		UserSvc: userSvc,
		CardSvc: cardSvc,
		DataSvc: dataSvc,
	})

	slog.Info("running server", "host", opts.RunAddr)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
