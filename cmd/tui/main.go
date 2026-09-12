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

type tabType uint

const (
	tabProfile tabType = iota
	tabCards
	tabData
)

func (t tabType) String() string {
	switch t {
	case tabProfile:
		return "profile"
	case tabCards:
		return "cards"
	case tabData:
		return "data"
	default:
		return ""
	}
}

type pageType uint

const (
	homePage pageType = iota
	registerPage
	loginPage
	confirmPage
	userPage
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
	case userPage:
		return "user"
	default:
		return ""
	}
}

type authForm struct {
	email   textinput.Model
	confirm textinput.Model
	pass    textinput.Model
}

type cardForm struct {
	number textinput.Model
	date   textinput.Model
}

type model struct {
	cursor            int
	activePage        pageType
	activeTab         tabType
	history           []pageType
	choices           []pageType
	allTabs           []tabType
	contentTabProfile []string
	authForm          authForm
	cardForm          cardForm
	cardFormActive    bool
	width             int
	height            int
	textError         string
	errorSeq          int
	authUser          string
	codeEmail         string
	accessToken       string
	refreshToken      string
	user              map[string]string
	cards             []*pb.CardData
	isLoad            bool
	spinner           spinner.Model

	client pb.GoKeeperServiceClient
}

func NewModel(client pb.GoKeeperServiceClient) model {
	email := newStyledInput("email@example.com", false)
	email.Focus()

	pass := newStyledInput("password", true)
	confirm := newStyledInput("confirmation code", false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	number := newStyledInput("card number", false)
	date := newStyledInput("09/26", false)

	return model{
		activePage:        homePage,
		activeTab:         tabProfile,
		history:           []pageType{homePage},
		choices:           []pageType{loginPage, registerPage},
		allTabs:           []tabType{tabProfile, tabCards, tabData},
		contentTabProfile: []string{"email", "password", "create_date"},
		authForm: authForm{
			email:   email,
			pass:    pass,
			confirm: confirm,
		},
		cardForm: cardForm{
			number: number,
			date:   date,
		},
		spinner: s,
		cards:   []*pb.CardData{},
		client:  client,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
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

	// обработка ответа API
	case registerResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err)
		}
		m.codeEmail = msg.emailCode
		m.activePage = confirmPage
		m.isLoad = false

	case loginResultMsg:
		if msg.err != nil {
			st, _ := status.FromError(msg.err)
			if st.Code() == codes.PermissionDenied {
				m.codeEmail = msg.emailCode
				m.activePage = confirmPage
			} else {
				m.textError = msg.err.Error()
			}
			m.isLoad = false
			return m, clearErrorAfter(m.errorSeq)
		}
		m.accessToken = msg.accessToken
		m.refreshToken = msg.refreshToken
		m.authUser = m.authForm.email.Value()
		m = m.navigateTo(userPage)
		return m, getProfileCmd(m.client, m.accessToken)

	case confirmResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err)
		}
		m.isLoad = true
		return m, loginCmd(m.client, m.authForm.email.Value(), m.authForm.pass.Value())

	case profileResultMsg:
		if msg.err != nil {
			m.activePage = homePage
			return m, m.handleError(msg.err)
		}
		m.user = msg.user
		m.isLoad = false

	case cardsResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err)
		}
		m.cards = msg.cards
		m.isLoad = false

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.Reset()
			return m, nil
		case "esc":
			if m.authUser != "" &&
				m.history[len(m.history)-1] == loginPage {
				m.textError = "to log out, press ctrl+q"
				return m, clearErrorAfter(m.errorSeq)
			}
			if len(m.history) > 1 {
				m.history = m.history[:len(m.history)-1]
				m.activePage = m.history[len(m.history)-1]
			}
			return m, nil
		}

		switch m.activePage {
		case registerPage:
			return registerUpdate(msg, m)
		case confirmPage:
			return confirmUpdate(msg, m)
		case loginPage:
			return loginUpdate(msg, m)
		case homePage:
			return homeUpdate(msg, m)
		case userPage:
			return userUpdate(msg, m)
		}

	default:
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

		switch {
		case m.authForm.email.Focused():
			m.authForm.email, cmd = m.authForm.email.Update(msg)
		case m.authForm.pass.Focused():
			m.authForm.pass, cmd = m.authForm.pass.Update(msg)
		case m.authForm.confirm.Focused():
			m.authForm.confirm, cmd = m.authForm.confirm.Update(msg)
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m model) View() tea.View {
	var content string
	s := headerView(m) + "\n"
	s += dividerView() + lipgloss.NewStyle().Render("\n")

	s += mainView(m)

	s += dividerView() + "\n"
	s += footerView()

	content = s
	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	return tea.NewView(centered)
}

func (m *model) handleError(err error) tea.Cmd {
	m.textError = err.Error()
	m.isLoad = false

	return clearErrorAfter(m.errorSeq)
}

func (m *model) Reset() {
	m.activePage = homePage
	m.history = []pageType{homePage}

	m.authUser = ""
	m.accessToken = ""
	m.refreshToken = ""
	m.codeEmail = ""
	m.cards = []*pb.CardData{}

	m.authForm.email.Focus()
	m.authForm.pass.Blur()
	m.authForm.confirm.Blur()
	m.authForm.email.SetValue("")
	m.authForm.pass.SetValue("")
	m.authForm.confirm.SetValue("")
}

type clearErrorMsg struct {
	seq int
}

func clearErrorAfter(seq int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{seq: seq}
	})
}

func (m model) navigateTo(activePage pageType) model {
	m.activePage = activePage
	m.history = append(m.history, activePage)

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
