package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

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
	page    pageType
	history []pageType
	choices []pageType
	cursor  int

	form      Form
	width     int
	height    int
	textError string
	errorSeq  int

	authUser     string
	codeEmail    string
	accessToken  string
	refreshToken string

	isLoad  bool
	spinner spinner.Model

	client pb.GoKeeperServiceClient
}

func NewModel(client pb.GoKeeperServiceClient) model {
	email := newStyledInput("email@example.com", false)
	email.Focus()

	pass := newStyledInput("password", true)
	confirm := newStyledInput("confirmation code", false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		client:  client,
		page:    homePage,
		history: []pageType{homePage},
		choices: []pageType{registerPage, loginPage},
		form: Form{
			email:   email,
			pass:    pass,
			confirm: confirm,
		},
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

		// очистка сообщения об ошибки
	case clearErrorMsg:
		if msg.seq == m.errorSeq {
			m.textError = ""
		}
		return m, nil

	// обработка ответа API(успех или ошибка)
	case resultMsg:
		if msg.err != nil {
			m.textError = msg.err.Error()

			st, _ := status.FromError(msg.err)
			if st.Code() == codes.PermissionDenied {
				m.codeEmail = msg.emailCode
				m.form.confirm.SetValue("")
				m.form.email.Blur()
				m.form.pass.Blur()
				m.form.confirm.Focus()
				m.page = confirmPage
			}
			return m, clearErrorAfter(m.errorSeq)
		}
		switch msg.kind {
		case "register":
			m.codeEmail = msg.emailCode
			m.page = confirmPage
		case "confirm":
			return m, loginCmd(m.client, m.form.email.Value(), m.form.pass.Value())
		case "login":
			m.accessToken = msg.accessToken
			m.refreshToken = msg.refreshToken
			m.authUser = m.form.email.Value()
			m = m.navigateTo(profilePage)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.page = homePage
			m.authUser = ""
			m.accessToken = ""
			return m, nil
		case "esc":
			if m.authUser != "" {
				m.textError = "to log out, press ctrl+q"
				return m, clearErrorAfter(m.errorSeq)
			}
			if len(m.history) > 1 {
				m.history = m.history[:len(m.history)-1]
				m.page = m.history[len(m.history)-1]
			}
			return m, nil
		}

		switch m.page {
		case registerPage:
			return registerUpdate(msg, m)
		case confirmPage:
			return confirmUpdate(msg, m)
		case loginPage:
			return loginUpdate(msg, m)
		case homePage:
			return homeUpdate(msg, m)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var content string
	s := headerView(cardWidth, "Go Keeper", breadcrumbView(m.page.String(), m.authUser)) + "\n"
	s += dividerView(cardWidth) + lipgloss.NewStyle().MarginBottom(2).Render("\n")

	s += mainView(m)

	s += dividerView(cardWidth) + "\n"
	s += footerView(cardWidth, "ctrl+c quit· ctrl+q logout · esc back · tab focus", "© Albert Taygibov")

	content = s
	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	return tea.NewView(centered)
}

type clearErrorMsg struct {
	seq int
}

func clearErrorAfter(seq int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{seq: seq}
	})
}

func (m model) navigateTo(page pageType) model {
	m.page = page
	m.history = append(m.history, page)

	return m
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

	p := tea.NewProgram(NewModel(c))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
