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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

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

	contentTabProfile []string
	cardsActions      []cardActionsType
	dataActions       []dataActionsType

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

	profile            map[string]string
	cards              []*pb.CardData
	arbitraryData      []*pb.ArbitraryData
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
		activePage:        initPage,
		history:           initHistory,
		activeTab:         tabProfile,
		choices:           []pageType{loginPage, registerPage},
		allTabs:           []tabType{tabProfile, tabCards, tabData, tabLocalData},
		contentTabProfile: []string{"email", "create_date"},
		cardsActions:      []cardActionsType{cardActionsUpdate, cardActionsDelete},
		dataActions:       []dataActionsType{dataActionsDownload, dataActionsReload, dataActionsDelete},
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.progress.SetWidth(60)
		if m.progress.Width() > maxWidth {
			m.progress.SetWidth(maxWidth)
		}
		m.width = msg.Width
		m.height = msg.Height
		var cmd tea.Cmd
		m.filepicker, cmd = m.filepicker.Update(msg)
		return m, cmd

		// очистка сообщения об ошибки
	case clearErrorMsg:
		m.textError = ""
		m.textInfo = ""
		return m, nil

		// API result public
	case registerResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "registerResultMsg")
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
			}
			return m, m.handleError(msg.err, "loginResultMsg")
		}
		m.accessToken = msg.accessToken
		m.refreshToken = msg.refreshToken
		m.authUser = m.authForm.email.Value()
		if err := m.localStorage.Set(emailKey, m.authForm.email.Value()); err != nil {
			return m, m.handleError(err, "loginResultMsg")
		}
		if err := m.localStorage.Set(accessTokenKey, msg.accessToken); err != nil {
			return m, m.handleError(err, "loginResultMsg")
		}
		if err := m.localStorage.Set(refreshTokenKey, msg.refreshToken); err != nil {
			return m, m.handleError(err, "loginResultMsg")
		}
		return m.navigateTo(homePage), getProfileCmd(m.client)

	case confirmResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "confirmResultMsg")
		}

		m.isLoad = false
		m.authForm.email.Focus()
		m.authForm.pass.Blur()
		m.authForm.pass.SetValue("")

		return m.navigateTo(loginPage), nil

	case refreshTokenErrorMsg:
		m.Reset()
		return m.navigateTo(loginPage), m.handleError(msg.err, "refreshTokenErrorMsg")

		// API result private
	case profileResultMsg:
		if msg.err != nil {
			st, _ := status.FromError(msg.err)
			if st.Code() == codes.Unavailable {
				m.Reset()
				return m, m.authForm.email.Focus()
			}
			return m, m.handleError(msg.err, "profileResultMsg")
		}
		m.profile = msg.profile
		m.isLoad = false

	case cardsResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "cardsResultMsg")
		}
		m.cards = msg.cards
		m.isLoad = false

	case createCardResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "createCardResultMsg")
		}

		m.cardForm.number.SetValue("")
		m.cardForm.date.SetValue("")
		m.cardForm.number.Blur()
		m.cardForm.date.Blur()
		// вызов getCardsCmd для получения нового списка после добавления
		return m.navigateTo(homePage), getCardsCmd(m.client)

	case changePassResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "changePassResultMsg")
		}
		m.isLoad = false
		return m.navigateTo(homePage), getProfileCmd(m.client)

	case deleteCardResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "deleteCardResultMsg")
		}
		m.isLoad = false
		m.selectedRowID = -1
		return m.navigateTo(homePage), getCardsCmd(m.client)

	case activateCardResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "activateCardResultMsg")
		}
		m.isLoad = false
		m.selectedRowID = -1
		return m.navigateTo(homePage), getCardsCmd(m.client)

	case arbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "arbitraryDataResultMsg")
		}
		m.arbitraryData = msg.arbitraryData
		m.isLoad = false

	case localArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "arbitraryDataResultMsg")
		}
		m.localArbitraryData = msg.localArbitraryData
		m.isLoad = false

	case createArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "createArbitraryDataResultMsg")
		}
		m.isLoad = false
		m.selectedFile = ""
		return m.navigateTo(homePage), getArbitraryDataCmd(m.client)

	case deleteArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "deleteArbitraryDataResultMsg")
		}
		m.isLoad = false
		m.selectedRowID = -1
		m.cursor = 0
		return m.navigateTo(homePage), getArbitraryDataCmd(m.client)

	case reloadArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "reloadArbitraryDataResultMsg")
		}
		m.isLoad = false
		m.selectedFile = ""
		m.selectedRowID = -1
		return m, getArbitraryDataCmd(m.client)

	case downloadArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "downloadArbitraryDataResultMsg")
		}
		m.isLoad = false
		m.selectedRowID = -1
		return m, nil

	case uploadProgressMsg:
		m.uploadedBytes = msg.sent
		m.totalBytes = msg.total

		return m, waitUploadProgress(m.progressCh)
		// Обработка нажатия клавиш
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.Reset()
			cmd := m.authForm.email.Focus()
			return m, cmd
		case "esc":
			if m.selectedFile != "" {
				m.selectedFile = ""
				return m, nil
			}
			if m.selectedRowID >= 0 {
				m.selectedRowID = -1
				return m, nil
			}
			if m.activePage == cardFormPage {
				m.cardForm.number.Blur()
				m.cardForm.date.Blur()
			}
			if len(m.history) <= 1 {
				return m, m.handleInfo("to log out, press ctrl+q", "")
			}
			if len(m.history) > 1 {
				m.history = m.history[:len(m.history)-1]
				m.activePage = m.history[len(m.history)-1]
			}
			return m, nil
		}

		switch m.activePage {
		case landingPage:
			return m.landingPageUpdate(msg)
		case registerPage:
			return m.registerPageUpdate(msg)
		case loginPage:
			return m.loginPageUpdate(msg)
		case confirmPage:
			return m.confirmPageUpdate(msg)
		case homePage:
			return m.homePageUpdate(msg)
		case cardFormPage:
			return m.cardFormPageUpdate(msg)
		case profileFormPage:
			return m.profileFormPageUpdate(msg)
		case filePickerPage:
			return m.filePickerPageUpdate(msg)
		}

	}
	m, cmd := m.updateNonKeyComponents(msg)
	return m, cmd
}

func (m model) View() tea.View {
	var content string
	s := m.headerView() + "\n"
	s += dividerView() + lipgloss.NewStyle().Render("\n")

	s += m.mainView()

	s += dividerView() + "\n"
	s += footerView()

	content = s
	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	v := tea.NewView(centered)
	v.AltScreen = true
	return v
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
