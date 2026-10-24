package main

import (
	"context"

	tea "charm.land/bubbletea/v2"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type registerResultMsg struct {
	err       error
	emailCode string
}

type confirmResultMsg struct {
	err error
}

type loginResultMsg struct {
	err          error
	emailCode    string
	accessToken  string
	refreshToken string
}

type refreshTokenErrorMsg struct {
	err error
}

func registerCmd(client pb.GoKeeperServiceClient, email, pass string) tea.Cmd {
	return func() tea.Msg {
		var header metadata.MD

		_, err := client.Register(
			context.Background(),
			pb.RegisterRequest_builder{Email: email, Pass: pass}.Build(),
			grpc.Header(&header),
		)
		if err != nil {
			return registerResultMsg{err: err}
		}

		var emailCode string
		if values := header.Get("email_code"); len(values) > 0 {
			emailCode = values[0]
		}
		return registerResultMsg{emailCode: emailCode}
	}
}

func confirmEmailCmd(client pb.GoKeeperServiceClient, email, code string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ConfirmEmail(
			context.Background(),
			pb.ConfirmEmailRequest_builder{Email: email, EmailCode: code}.Build(),
		)
		return confirmResultMsg{err: err}
	}
}

func loginCmd(client pb.GoKeeperServiceClient, email, pass string) tea.Cmd {
	return func() tea.Msg {
		var header metadata.MD

		resp, err := client.Login(
			context.Background(),
			pb.LoginRequest_builder{Email: email, Pass: pass}.Build(),
			grpc.Header(&header),
		)
		if err != nil {
			var emailCode string
			if values := header.Get("email_code"); len(values) > 0 {
				emailCode = values[0]
			}
			return loginResultMsg{err: err, emailCode: emailCode}
		}
		return loginResultMsg{
			accessToken:  resp.GetAccessToken(),
			refreshToken: resp.GetRefreshToken(),
		}
	}
}
