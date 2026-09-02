package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
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

	in := bufio.NewReader(os.Stdin)

	fmt.Println("  register email password - регистрация")
	fmt.Println("  login email password    - вход")
	fmt.Println("  exit                    - выход")

	for {
		fmt.Print("client > ")

		line, err := in.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения:", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		fmt.Println(parts)

		command := parts[0]

		switch command {
		case "register":
			if len(parts) < 3 {
				fmt.Println("Использование: register email password")
				continue
			}
			email := parts[1]
			password := parts[2]

			resp, err := c.Register(ctx, pb.RegisterRequest_builder{
				Email:    email,
				Password: password,
			}.Build())
			if err == nil {
				fmt.Println("Register OK", resp.String())
			}

		case "login":
			email := parts[1]
			password := parts[2]

			resp, err := c.Login(ctx, pb.LoginRequest_builder{
				Email:    email,
				Password: password,
			}.Build())
			if err == nil {
				fmt.Println("Login OK", resp.String())
			}

		default:
			fmt.Printf("Неизвестная команда: %s\n", command)
			fmt.Println("Доступные команды: register, login, exit")
		}

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
