package main

import (
	"context"
	"slices"

	tea "charm.land/bubbletea/v2"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

func getCardsCmd(client pb.GoKeeperServiceClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetCards(context.Background(), &pb.CardsRequest{})
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.Unauthenticated {
				return refreshTokenErrorMsg{err: err}
			}
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
			CardNumber: []byte(number),
			ExpiryDate: expiry,
		}.Build())
		st, _ := status.FromError(err)
		if st.Code() == codes.Unauthenticated {
			return refreshTokenErrorMsg{err: err}
		}
		return createCardResultMsg{err: err}
	}
}

func deleteCardCmd(client pb.GoKeeperServiceClient, id int64) tea.Cmd {
	return func() tea.Msg {
		_, err := client.DeleteCard(context.Background(), pb.DeleteCardRequest_builder{
			Id: id,
		}.Build())
		st, _ := status.FromError(err)
		if st.Code() == codes.Unauthenticated {
			return refreshTokenErrorMsg{err: err}
		}
		return deleteCardResultMsg{err: err}
	}
}

func activateCardCmd(client pb.GoKeeperServiceClient, id int64) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ActivateCard(context.Background(), pb.ActiveCardRequest_builder{
			Id: id,
		}.Build())
		st, _ := status.FromError(err)
		if st.Code() == codes.Unauthenticated {
			return refreshTokenErrorMsg{err: err}
		}
		return activateCardResultMsg{err: err}
	}
}
