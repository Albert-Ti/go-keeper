package main

import (
	"context"

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

func (a *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	if !publicMethods[method] {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", a.localStorage.Get(accessTokenKey))
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (a *AuthInterceptor) StreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	if !publicMethods[method] {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", a.localStorage.Get(accessTokenKey))
	}
	return streamer(ctx, desc, cc, method, opts...)
}
