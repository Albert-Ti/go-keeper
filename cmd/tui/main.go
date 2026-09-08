package main

import (
	"fmt"
	"log/slog"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

type pageType uint

const (
	homePage pageType = iota
	registerPage
	loginPage
	confirmPage
	profilePage
)

func (p pageType) String() string {
	switch p {
	case homePage:
		return "home"
	case registerPage:
		return "registration"
	case loginPage:
		return "login"
	case confirmPage:
		return "confirm"
	case profilePage:
		return "profile"
	default:
		return ""
	}
}

type Form struct {
	email   textinput.Model
	confirm textinput.Model
	pass    textinput.Model
}

type model struct {
	page         pageType
	history      []pageType
	choices      []string
	cursor       int
	form         Form
	authUser     string
	textError    string
	errorSeq     int
	codeEmail    string
	accessToken  string
	refreshToken string
	width        int
	height       int

	client pb.GoKeeperServiceClient
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func initialModel(client pb.GoKeeperServiceClient) model {
	email := newStyledInput("email@example.com", false)
	email.Focus()

	pass := newStyledInput("password", true)
	confirm := newStyledInput("confirmation code", false)

	return model{
		client:  client,
		page:    homePage,
		history: []pageType{homePage},
		choices: []string{"register", "login"},
		form: Form{
			email:   email,
			pass:    pass,
			confirm: confirm,
		},
	}
}

func main() {
	conn, err := grpc.NewClient(
		"127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("ошибка при установлении соединения с сервером", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	c := pb.NewGoKeeperServiceClient(conn)

	p := tea.NewProgram(initialModel(c))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
