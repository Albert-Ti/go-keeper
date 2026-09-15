package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var publicMethods = map[string]bool{
	"/gokeeper.GoKeeperService/Register":     true,
	"/gokeeper.GoKeeperService/Login":        true,
	"/gokeeper.GoKeeperService/ConfirmEmail": true,
	"/gokeeper.GoKeeperService/RefreshToken": true,
}

type AuthInterceptor struct {
	localStorage *FileStorage
}

func (a *AuthInterceptor) UnaryInterceptor(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if !publicMethods[method] {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", a.localStorage.creds.AccessToken)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
