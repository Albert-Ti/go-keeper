package main

import (
	"context"
	"time"

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

type profileResultMsg struct {
	err  error
	user map[string]string
}

type cardsResultMsg struct {
	err   error
	cards []*pb.CardData
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
			pb.LoginRequest_builder{Email: email, Password: pass}.Build(),
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

func getProfileCmd(client pb.GoKeeperServiceClient, token string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

		resp, err := client.GetProfile(ctx, &pb.ProfileRequest{})
		if err != nil {
			return profileResultMsg{err: err}
		}
		return profileResultMsg{err: err, user: map[string]string{
			"email":       resp.GetEmail(),
			"password":    "*******",
			"create_date": resp.GetCreatedAt().AsTime().Format("02 Jan 2006, 15:04"),
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
			return cardsResultMsg{err: err}
		}
		return cardsResultMsg{err: err, cards: resp.GetCards()}
	}
}
