package main

import (
	tea "charm.land/bubbletea/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.Reset()
			cmd := m.authForm.email.Focus()
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.progress.SetWidth(60)
		if m.progress.Width() > maxWidth {
			m.progress.SetWidth(maxWidth)
		}
		m.width = msg.Width
		m.height = msg.Height

		// очистка сообщения об ошибки
	case clearErrorMsg:
		m.textError = ""
		m.textInfo = ""

		// API result public
	case registerResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "registerResultMsg")
		}

		m.codeEmail = msg.emailCode
		m.activePage = confirmPage
		m.isLoad = false
		m.authForm.email.Blur()
		m.authForm.pass.Blur()
		cmd := m.authForm.confirm.Focus()

		return m, cmd

	case loginResultMsg:
		if msg.err != nil {
			st, _ := status.FromError(msg.err)
			if st.Code() == codes.PermissionDenied {
				m.codeEmail = msg.emailCode
				m.activePage = confirmPage
				m.authForm.email.Blur()
				m.authForm.pass.Blur()
				cmd := m.authForm.confirm.Focus()
				return m, cmd
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
		cmd := m.authForm.email.Focus()
		return m.navigateTo(loginPage), tea.Batch(
			m.handleError(msg.err, "refreshTokenErrorMsg"), cmd)

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
		return m.navigateTo(homePage), getArbitraryDataCmd(m.client, m.localStorage.Get(masterKeyKey))

	case deleteArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "deleteArbitraryDataResultMsg")
		}
		m.isLoad = false
		m.selectedRowID = -1
		m.cursor = 0
		return m.navigateTo(homePage), getArbitraryDataCmd(m.client, m.localStorage.Get(masterKeyKey))

	case reloadArbitraryDataResultMsg:
		if msg.err != nil {
			return m, m.handleError(msg.err, "reloadArbitraryDataResultMsg")
		}
		m.isLoad = false
		m.selectedFile = ""
		m.selectedRowID = -1
		return m, getArbitraryDataCmd(m.client, m.localStorage.Get(masterKeyKey))

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

	return m, nil
}
