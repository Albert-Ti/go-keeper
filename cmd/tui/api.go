package main

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type resultMsg struct {
	err  error
	kind string

	emailCode string

	accessToken  string
	refreshToken string

	cards []*pb.CardData
	user  map[string]string
}

func registerCmd(client pb.GoKeeperServiceClient, email, pass string) tea.Cmd {
	return func() tea.Msg {
		var header metadata.MD

		_, err := client.Register(
			context.Background(),
			pb.RegisterRequest_builder{Email: email, Password: pass}.Build(),
			grpc.Header(&header),
		)
		if err != nil {
			return resultMsg{err: err, kind: "register"}
		}

		var emailCode string
		if values := header.Get("email_code"); len(values) > 0 {
			emailCode = values[0]
		}
		return resultMsg{kind: "register", emailCode: emailCode}
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
		return resultMsg{
			kind:         "login",
			accessToken:  resp.GetAccessToken(),
			refreshToken: resp.GetRefreshToken(),
		}
	}
}

func getProfileCmd(client pb.GoKeeperServiceClient, token string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

		resp, err := client.GetProfile(ctx, &pb.ProfileRequest{})
		if err != nil {
			return resultMsg{err: err, kind: "profile"}
		}
		return resultMsg{err: err, kind: "profile", user: map[string]string{
			"email":    resp.GetEmail(),
			"password": "*******",
			"date":     resp.GetCreatedAt(),
		}}
	}
}

func getCardsCmd(client pb.GoKeeperServiceClient, token string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

		resp, err := client.GetCards(ctx, &pb.CardsRequest{})
		if err != nil {
			return resultMsg{err: err, kind: "cards"}
		}
		return resultMsg{err: err, kind: "cards", cards: resp.GetCards()}
	}
}
