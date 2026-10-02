package main

import (
	"context"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
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

type refreshTokenResultMsg struct {
	err          error
	accessToken  string
	refreshToken string
}

type changePassResultMsg struct {
	err error
}

type profileResultMsg struct {
	err     error
	profile map[string]string
}

type cardsResultMsg struct {
	err   error
	cards []*pb.CardData
}

type createCardResultMsg struct {
	err error
}

type deleteCardResultMsg struct {
	err error
}

type activateCardResultMsg struct {
	err error
}

type arbitraryDataResultMsg struct {
	err           error
	arbitraryData []*pb.ArbitraryData
}
type createArbitraryDataResultMsg struct {
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

func refreshTokenCmd(client pb.GoKeeperServiceClient, token string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.RefreshToken(ctx, pb.TokenRequest_builder{
			RefreshToken: token,
		}.Build())
		if err != nil {
			return refreshTokenResultMsg{err: err}
		}
		return refreshTokenResultMsg{
			accessToken:  resp.GetAccessToken(),
			refreshToken: resp.GetRefreshToken(),
		}
	}
}

func changePassCmd(client pb.GoKeeperServiceClient, passOld, passNew string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ChangePass(context.Background(), pb.PassRequest_builder{
			PassOld: passOld,
			PassNew: passNew,
		}.Build())
		return changePassResultMsg{err: err}
	}
}

func getProfileCmd(client pb.GoKeeperServiceClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetProfile(context.Background(), &pb.ProfileRequest{})
		if err != nil {
			return profileResultMsg{err: err}
		}
		return profileResultMsg{err: err, profile: map[string]string{
			"email":       resp.GetEmail(),
			"pass":        "*******",
			"create_date": resp.GetCreatedAt().AsTime().Format("02 Jan 2006, 15:04"),
		}}
	}
}

func getCardsCmd(client pb.GoKeeperServiceClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetCards(context.Background(), &pb.CardsRequest{})
		if err != nil {
			return cardsResultMsg{err: err}
		}

		slices.SortFunc(resp.GetCards(), func(a, b *pb.CardData) int {
			aActive := a.GetActive()
			bActive := b.GetActive()
			switch {
			case aActive && !bActive:
				return -1
			case !aActive && bActive:
				return 1
			default:
				return 0
			}
		})
		return cardsResultMsg{err: err, cards: resp.GetCards()}
	}
}

func createCardCmd(client pb.GoKeeperServiceClient, number, expiry string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.CreateCard(context.Background(), pb.CreateCardRequest_builder{
			CardNumber: number,
			ExpiryDate: expiry,
		}.Build())

		return createCardResultMsg{err: err}
	}
}

func deleteCardCmd(client pb.GoKeeperServiceClient, id int64) tea.Cmd {
	return func() tea.Msg {
		_, err := client.DeleteCard(context.Background(), pb.DeleteCardRequest_builder{
			Id: id,
		}.Build())

		return deleteCardResultMsg{err: err}
	}
}

func activateCardCmd(client pb.GoKeeperServiceClient, id int64) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ActivateCard(context.Background(), pb.ActiveCardRequest_builder{
			Id: id,
		}.Build())

		return activateCardResultMsg{err: err}
	}
}

func createArbitraryDataCmd(client pb.GoKeeperServiceClient, filepath string) tea.Cmd {
	return func() tea.Msg {
		var filename string
		if os.PathSeparator == '\\' {
			s := strings.Split(filepath, "\\")
			filename = s[len(s)-1]
		} else {
			s := strings.Split(filepath, "/")
			filename = s[len(s)-1]
		}

		file, err := os.Open(filepath)
		if err != nil {
			return createArbitraryDataResultMsg{err: errors.New("failed to open: " + err.Error())}
		}
		defer file.Close()

		stream, err := client.CreateArbitraryData(context.Background())

		s := strings.Split(filename, ".")
		filetype := s[len(s)-1]

		info, err := file.Stat()
		if err != nil {
			return createArbitraryDataResultMsg{err: err}
		}

		// отправляем метаданные первым сообщением
		err = stream.Send(pb.CreateArbitraryDataRequest_builder{
			Metadata: pb.FileMetadata_builder{
				Filename: filename,
				Type:     filetype,
				Size:     info.Size(),
			}.Build(),
		}.Build())
		if err != nil {
			return createArbitraryDataResultMsg{err: errors.New("failed to send metadata: " + err.Error())}
		}

		buf := make([]byte, 1024*1024) // 1mb

		for {
			n, readErr := file.Read(buf)

			if n > 0 {
				if err := stream.Send(
					pb.CreateArbitraryDataRequest_builder{
						Chunk: buf[:n],
					}.Build(),
				); err != nil {
					return createArbitraryDataResultMsg{err: err}
				}
			}

			if readErr == io.EOF {
				break
			}

			if readErr != nil {
				return createArbitraryDataResultMsg{err: readErr}
			}
		}

		_, err = stream.CloseAndRecv()
		if err != nil {
			return createArbitraryDataResultMsg{err: err}
		}

		return createArbitraryDataResultMsg{err: nil}
	}
}

func getArbitraryDataCmd(client pb.GoKeeperServiceClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetArbitraryData(context.Background(), &pb.ListArbitraryDataRequest{})
		if err != nil {
			return cardsResultMsg{err: err}
		}

		return arbitraryDataResultMsg{err: err, arbitraryData: resp.GetArbitraryData()}
	}
}
