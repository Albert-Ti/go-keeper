package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
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

type tabType int

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

type cardActionsType int

const (
	cardActionsUpdate cardActionsType = iota
	cardActionsDelete
)

type pageType int

const (
	homePage pageType = iota
	registerPage
	loginPage
	confirmPage
	userPage
	profileFormPage
	cardFormPage
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
	case profileFormPage:
		return "profile/update"
	case cardFormPage:
		return "cards/create"
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

type profileForm struct {
	passOld textinput.Model
	passNew textinput.Model
}

type model struct {
	cursor            int
	activePage        pageType
	activeTab         tabType
	history           []pageType
	choices           []pageType
	allTabs           []tabType
	cardsActions      []cardActionsType
	contentTabProfile []string

	authForm    authForm
	cardForm    cardForm
	profileForm profileForm

	width         int
	height        int
	textError     string
	errorSeq      int
	codeEmail     string
	authUser      string
	accessToken   string
	refreshToken  string
	profile       map[string]string
	cards         []*pb.CardData
	selectedCard  int
	activeCardBtn cardActionsType
	isLoad        bool
	spinner       spinner.Model
	localStorage  *FileStorage

	client pb.GoKeeperServiceClient
}

func NewModel(client pb.GoKeeperServiceClient, localStorage *FileStorage) (*model, error) {
	email := newStyledInput("email@example.com", false)
	email.SetValue(localStorage.Get(emailKey))
	email.Focus()

	pass := newStyledInput("password", true)
	passOld := newStyledInput("old password", true)
	passNew := newStyledInput("new password", true)
	confirm := newStyledInput("confirmation code", false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	number := newStyledInput("card number", false)
	date := newStyledInput("09/26", false)

	initPage := homePage
	initHistory := []pageType{homePage}
	if localStorage.Get(accessTokenKey) != "" {
		initPage = userPage
		initHistory = []pageType{userPage}
	}

	return &model{
		activePage:        initPage,
		activeTab:         tabProfile,
		history:           initHistory,
		choices:           []pageType{loginPage, registerPage},
		allTabs:           []tabType{tabProfile, tabCards, tabData},
		contentTabProfile: []string{"email", "create_date"},
		cardsActions:      []cardActionsType{cardActionsUpdate, cardActionsDelete},
		authForm: authForm{
			email:   email,
			pass:    pass,
			confirm: confirm,
		},
		cardForm: cardForm{
			number: number,
			date:   date,
		},
		profileForm: profileForm{
			passOld: passOld,
			passNew: passNew,
		},
		spinner:      s,
		cards:        []*pb.CardData{},
		client:       client,
		selectedCard: -1,
		localStorage: localStorage,
		authUser:     localStorage.Get(emailKey),
		accessToken:  localStorage.Get(accessTokenKey),
		refreshToken: localStorage.Get(refreshTokenKey),
	}, nil
}

func (m model) Init() tea.Cmd {
	if m.accessToken == "" {
		return tea.Batch(
			textinput.Blink,
			m.spinner.Tick,
		)
	}

	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		// Минуем авторизацию если есть токен
		getProfileCmd(m.client),
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

		// API result public
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
		if err := m.localStorage.Set(emailKey, m.authForm.email.Value()); err != nil {
			return m, m.handleError(err)
		}
		if err := m.localStorage.Set(accessTokenKey, msg.accessToken); err != nil {
			return m, m.handleError(err)
		}
		if err := m.localStorage.Set(refreshTokenKey, msg.refreshToken); err != nil {
			return m, m.handleError(err)
		}
		return m.navigateTo(userPage), getProfileCmd(m.client)

	case confirmResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err)
		}
		m.isLoad = false
		m.authForm.email.Focus()
		m.authForm.pass.Blur()
		m.authForm.pass.SetValue("")
		return m.navigateTo(loginPage), nil

	case refreshTokenResultMsg:
		if msg.err != nil {
			m.Reset()
			return m, m.handleError(msg.err)
		}
		m.accessToken = msg.accessToken
		m.refreshToken = msg.refreshToken
		if err := m.localStorage.Set(accessTokenKey, msg.accessToken); err != nil {
			return m, m.handleError(err)
		}
		if err := m.localStorage.Set(refreshTokenKey, msg.refreshToken); err != nil {
			return m, m.handleError(err)
		}
		return m, getProfileCmd(m.client)

		// API result private
	case profileResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "access token is expired") {
				return m, refreshTokenCmd(m.client, m.localStorage.Get(refreshTokenKey))
			}
			m.Reset()
			return m, m.handleError(msg.err)
		}
		m.profile = msg.profile
		m.isLoad = false

	case cardsResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "access token is expired") {
				return m, refreshTokenCmd(m.client, m.localStorage.Get(refreshTokenKey))
			}
			m.Reset()
			return m, m.handleError(msg.err)
		}
		m.cards = msg.cards
		m.isLoad = false

	case createCardResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "access token is expired") {
				return m, refreshTokenCmd(m.client, m.localStorage.Get(refreshTokenKey))
			}
			return m, m.handleError(msg.err)
		}

		m.cardForm.number.SetValue("")
		m.cardForm.date.SetValue("")
		m.cardForm.number.Blur()
		m.cardForm.date.Blur()
		// вызов getCardsCmd для получения нового списка после добавления
		return m.navigateTo(userPage), getCardsCmd(m.client)

	case changePassResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "access token is expired") {
				return m, refreshTokenCmd(m.client, m.localStorage.Get(refreshTokenKey))
			}
			return m, m.handleError(msg.err)
		}
		m.isLoad = false
		return m.navigateTo(userPage), getProfileCmd(m.client)

	case deleteCardResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "access token is expired") {
				return m, refreshTokenCmd(m.client, m.localStorage.Get(refreshTokenKey))
			}
			return m, m.handleError(msg.err)
		}
		m.isLoad = false
		m.selectedCard = -1
		return m.navigateTo(userPage), getCardsCmd(m.client)

	case activateCardResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "access token is expired") {
				return m, refreshTokenCmd(m.client, m.localStorage.Get(refreshTokenKey))
			}
			return m, m.handleError(msg.err)
		}
		m.isLoad = false
		m.selectedCard = -1
		return m.navigateTo(userPage), getCardsCmd(m.client)

		// Обработка нажатия клавиш
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.Reset()
			return m, nil
		case "esc":
			if m.selectedCard >= 0 {
				m.selectedCard = -1
				return m, nil
			}
			if m.activePage == cardFormPage {
				m.cardForm.number.Blur()
				m.cardForm.date.Blur()
			}
			if len(m.history) <= 1 {
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
		case cardFormPage:
			return cardFormUpdate(msg, m)
		case profileFormPage:
			return profileFormUpdate(msg, m)
		}

	default:
		var (
			cmd  tea.Cmd
			cmds []tea.Cmd
		)
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		// курсор в полях чтобы мигал
		switch {
		case m.authForm.email.Focused():
			m.authForm.email, cmd = m.authForm.email.Update(msg)
		case m.authForm.pass.Focused():
			m.authForm.pass, cmd = m.authForm.pass.Update(msg)
		case m.authForm.confirm.Focused():
			m.authForm.confirm, cmd = m.authForm.confirm.Update(msg)

		case m.cardForm.number.Focused():
			m.cardForm.number, cmd = m.cardForm.number.Update(msg)
		case m.cardForm.date.Focused():
			m.cardForm.date, cmd = m.cardForm.date.Update(msg)

		case m.profileForm.passOld.Focused():
			m.profileForm.passOld, cmd = m.profileForm.passOld.Update(msg)
		case m.profileForm.passNew.Focused():
			m.profileForm.passNew, cmd = m.profileForm.passNew.Update(msg)
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
	m.activeTab = tabProfile
	m.history = []pageType{homePage}

	m.authUser = ""
	m.accessToken = ""
	m.refreshToken = ""
	m.codeEmail = ""
	m.selectedCard = -1
	m.authForm.email.Focus()
	m.authForm.pass.Blur()
	m.authForm.confirm.Blur()
	m.authForm.email.SetValue(m.localStorage.Get(emailKey))
	m.authForm.pass.SetValue("")
	m.authForm.confirm.SetValue("")

	m.cardForm.number.Focus()
	m.cardForm.number.Blur()
	m.cardForm.date.Blur()
	m.cardForm.number.SetValue("")
	m.cardForm.date.SetValue("")
}

type clearErrorMsg struct {
	seq int
}

func clearErrorAfter(seq int) tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{seq: seq}
	})
}

func (m model) navigateTo(activePage pageType) model {
	if activePage == homePage || activePage == userPage {
		m.history = m.history[:0]
	}
	m.activePage = activePage
	m.history = append(m.history, activePage)

	return m
}

func main() {
	logger, closeLog, err := setupLogger("client.log")
	if err != nil {
		fmt.Println("setup logger fatal:", err)
		os.Exit(1)
	}
	defer closeLog()

	slog.SetDefault(logger) // slog.Info/Error/Debug пишут в файл везде

	localStorage, err := NewFileStorage()
	if err != nil {
		panic(err)
	}

	auth := AuthInterceptor{
		localStorage: localStorage,
	}
	conn, err := grpc.NewClient(
		"127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(auth.UnaryInterceptor),
	)
	if err != nil {
		slog.Error("ошибка при установлении соединения с сервером", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	c := pb.NewGoKeeperServiceClient(conn)

	model, err := NewModel(c, localStorage)
	if err != nil {
		panic(err)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
