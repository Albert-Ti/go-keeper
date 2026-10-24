package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Albert-Ti/go-keeper/pkg/proto"
)

type model struct {
	isLoad bool

	cursor        int
	selectedRowID int
	activeBtn     int

	activePage pageType
	activeTab  tabType
	history    []pageType
	choices    []pageType
	allTabs    []tabType

	cardsActions []cardActionsType
	dataActions  []dataActionsType

	authForm    authForm
	cardForm    cardForm
	profileForm profileForm

	spinner    spinner.Model
	filepicker filepicker.Model

	progress      progress.Model
	uploadedBytes int64
	totalBytes    int64
	progressCh    chan uploadProgressMsg

	selectedFile string
	width        int
	height       int
	textError    string
	textInfo     string
	codeEmail    string
	authUser     string
	accessToken  string
	refreshToken string

	localStorage *FileStorage

	profile            profileUser
	cards              []*pb.CardData
	arbitraryData      []arbitraryData
	localArbitraryData []localUploadedData
	privateData        bool

	keyPressInfo []string

	client pb.GoKeeperServiceClient
}

func NewModel(client pb.GoKeeperServiceClient, localStorage *FileStorage) (*model, error) {
	email := newStyledInput("email@example.com", false)
	email.SetValue(localStorage.Get(emailKey))

	pass := newStyledInput("password", true)
	passCurr := newStyledInput("current password", true)
	passNew := newStyledInput("new password", true)
	confirm := newStyledInput("confirmation code", false)
	number := newStyledInput("card number", false)
	date := newStyledInput("09/26", false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorAccent)
	prog := progress.New(progress.WithScaled(true), progress.WithColors(pink, yellow))

	fp := filepicker.New()
	fp.AllowedTypes = []string{".html", ".txt", ".jpg", ".png", ".mp4", ".mkv", ".mov"}
	var err error
	fp.CurrentDirectory, err = os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fp.AutoHeight = false
	fp.SetHeight(6)

	initPage := landingPage
	initHistory := []pageType{landingPage}
	if localStorage.Get(accessTokenKey) != "" {
		initPage = homePage
		initHistory = []pageType{homePage}
	}

	return &model{
		activePage:   initPage,
		history:      initHistory,
		activeTab:    tabProfile,
		choices:      []pageType{loginPage, registerPage},
		allTabs:      []tabType{tabProfile, tabCards, tabData, tabLocalData},
		cardsActions: []cardActionsType{cardActionsUpdate, cardActionsDelete},
		dataActions:  []dataActionsType{dataActionsDownload, dataActionsReload, dataActionsDelete},
		privateData:  localStorage.Get(masterKeyKey) != "",
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
			passCurr: passCurr,
			passNew:  passNew,
		},
		spinner:    s,
		filepicker: fp,
		progress:   prog,
		progressCh: make(chan uploadProgressMsg, 10),

		cards:         []*pb.CardData{},
		client:        client,
		selectedRowID: -1,
		localStorage:  localStorage,
		authUser:      localStorage.Get(emailKey),
		accessToken:   localStorage.Get(accessTokenKey),
		refreshToken:  localStorage.Get(refreshTokenKey),
		keyPressInfo: []string{
			"ctrl+c quit",
			"ctrl+q logout",
			"esc back",
			"tab focus",
			"enter select",
			"↑↓←→ navigation",
		},
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
		m.filepicker.Init(),
		textinput.Blink,
		m.spinner.Tick,
		// Минуем авторизацию если есть токен
		getProfileCmd(m.client),
	)
}

func (m *model) handleError(err error, trigger string) tea.Cmd {
	m.textError = err.Error()
	m.isLoad = false
	if trigger != "" {
		slog.Error(trigger, "value_error", m.textError)
	}

	return clearErrorAfter()
}

func (m *model) handleInfo(info, trigger string) tea.Cmd {
	m.textInfo = info
	m.isLoad = false
	if trigger != "" {
		slog.Info(trigger, "value_info", info)
	}

	return clearErrorAfter()
}

func (m *model) Reset() {
	m.cursor = 0
	m.activePage = landingPage
	m.activeTab = tabProfile
	m.history = []pageType{landingPage}

	m.authUser = ""
	m.accessToken = ""
	m.refreshToken = ""
	m.codeEmail = ""
	m.selectedRowID = -1
	m.activeBtn = 0

	m.authForm.email.SetValue(m.localStorage.Get(emailKey))
	m.authForm.pass.SetValue("")
	m.authForm.confirm.SetValue("")
	m.cardForm.number.SetValue("")
	m.cardForm.date.SetValue("")
	m.profileForm.passCurr.SetValue("")
	m.profileForm.passNew.SetValue("")
}

type clearErrorMsg struct {
}

func clearErrorAfter() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func waitUploadProgress(ch <-chan uploadProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func (m model) navigateTo(activePage pageType) model {
	if activePage == landingPage || activePage == homePage {
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
	slog.SetDefault(logger)

	localStorage, err := NewFileStorage()
	if err != nil {
		panic(err)
	}

	auth := NewAuthInterceptor(localStorage)

	conn, err := grpc.NewClient(
		"127.0.0.1:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(auth.UnaryInterceptor),
		grpc.WithChainStreamInterceptor(auth.StreamInterceptor),
	)
	if err != nil {
		slog.Error("ошибка при установлении соединения с сервером", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewGoKeeperServiceClient(conn)

	model, err := NewModel(client, localStorage)
	if err != nil {
		panic(err)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
