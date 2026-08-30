package handler_test

import (
	"context"
	"net"
	"testing"

	"github.com/Albert-Ti/go-keeper/internal/config"
	"github.com/Albert-Ti/go-keeper/internal/handler"
	"github.com/Albert-Ti/go-keeper/internal/interceptor"
	"github.com/Albert-Ti/go-keeper/internal/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

const bufSize = 1024 * 1024

func newTestGRPCServer(t *testing.T, svc *service.Service, opts *config.Options) pb.GoKeeperServiceClient {
	t.Helper()

	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.AuthGuard(opts.JWTSecret),
			interceptor.Logging(),
		),
	)
	gs := &handler.GrpcServer{Srv: srv, Svc: svc}
	pb.RegisterGoKeeperServiceServer(srv, gs)

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("bufconn server exited: %v", err)
		}
	}()
	t.Cleanup(srv.Stop)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return pb.NewGoKeeperServiceClient(conn)
}
