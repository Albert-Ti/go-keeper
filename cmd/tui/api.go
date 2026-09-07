package main

import (
	"context"

	tea "charm.land/bubbletea/v2"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type resultMsg struct {
	err          error
	kind         string
	emailCode    string
	accessToken  string
	refreshToken string
}

func registerCmd(client pb.GoKeeperServiceClient, email, pass string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.Register(
			context.Background(),
			pb.RegisterRequest_builder{Email: email, Password: pass}.Build(),
		)
		if err != nil {
			return resultMsg{err: err, kind: "register"}
		}
		return resultMsg{kind: "register", emailCode: resp.GetConfirmCode()}
	}
}

func confirmEmailCmd(client pb.GoKeeperServiceClient, email, code string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ConfirmEmail(
			context.Background(),
			pb.ConfirmEmailRequest_builder{Email: email, EmailCode: code}.Build(),
		)
		return resultMsg{err: err, kind: "confirm"}
	}
}

func loginCmd(client pb.GoKeeperServiceClient, email, pass string) tea.Cmd {
	return func() tea.Msg {
		var header metadata.MD

		resp, err := client.Login(
			context.Background(),
			pb.LoginRequest_builder{Email: email, Password: pass}.Build(),
			grpc.Header(&header),
		)
		if err != nil {
			var emailCode string
			if values := header.Get("email_code"); len(values) > 0 {
				emailCode = values[0]
			}
			return resultMsg{err: err, kind: "login", emailCode: emailCode}

		}
		return resultMsg{kind: "login", accessToken: resp.GetAccessToken()}
	}
}
