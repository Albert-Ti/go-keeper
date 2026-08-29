package main

import (
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
