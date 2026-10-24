package main

import (
	"context"

	tea "charm.land/bubbletea/v2"
	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type changePassResultMsg struct {
	err error
}

type profileResultMsg struct {
	err     error
	profile profileUser
}

func changePassCmd(client pb.GoKeeperServiceClient, passCurr, passNew string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.ChangePass(context.Background(), pb.PassRequest_builder{
			PassCurr: passCurr,
			PassNew:  passNew,
		}.Build())
		return changePassResultMsg{err: err}
	}
}

func getProfileCmd(client pb.GoKeeperServiceClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetProfile(context.Background(), &pb.ProfileRequest{})
		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() == codes.Unauthenticated {
				return refreshTokenErrorMsg{err: err}
			}
			return profileResultMsg{err: err}
		}
		return profileResultMsg{err: err, profile: profileUser{
			email:     resp.GetEmail(),
			pass:      "*******",
			createdAt: resp.GetCreatedAt().AsTime().Format("02 Jan 2006, 15:04"),
		}}
	}
}
