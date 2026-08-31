package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

var currentToken string

func main() {
	ctx := context.Background()

	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(
		"127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientInterceptor),
	)

	if err != nil {
		slog.Error("ошибка при установлении соединения с сервером", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	c := pb.NewGoKeeperServiceClient(conn)

	resp, err := c.Register(ctx, pb.RegisterRequest_builder{
		Email:    "example@mail.com",
		Password: "12345",
	}.Build())
	if err == nil {
		fmt.Println("Register OK", resp.String())
	}

	resp2, err := c.Login(ctx, pb.LoginRequest_builder{
		Email:    "example@mail.com",
		Password: "12345",
	}.Build())
	if err == nil {
		fmt.Println("Register OK", resp2.String())
	}
}

func clientInterceptor(
	ctx context.Context, method string, req any, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	var header metadata.MD
	opts = append(opts, grpc.Header(&header)) // ловим заголовки ответа

	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)

	if values := header.Get("authorization"); len(values) > 0 {
		fmt.Println("получен новый токен:", values[0])
		currentToken = values[0]
	}

	if err != nil {
		log.Printf("[ERROR] %s,%v", method, err)
	} else {
		log.Printf("[INFO] %s,%v", method, time.Since(start))
	}
	return err
}
