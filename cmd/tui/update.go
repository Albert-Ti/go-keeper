package main

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type clearErrorMsg struct {
	seq int
}

func clearErrorAfter(seq int) tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{seq: seq}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

		// -------------- очистка сообщения об ошибки
	case clearErrorMsg:
		if msg.seq == m.errorSeq {
			m.textError = ""
		}
		return m, nil

	// -------------- обработка ответа API(успех или ошибка)
	case resultMsg:
		if msg.err != nil {
			m.textError = "⚠ " + msg.err.Error()

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
			m.token = msg.accessToken
			m.authUser = m.form.email.Value()
			m.page = profilePage
			m.history = append(m.history, profilePage)
		}
		return m, nil

	// -------------- обработка нажатия клавиш
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			m.page = homePage
			m.authUser = ""
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
		// -------------- обработка нажатия клавиш => Страница регистрации
		case registerPage:
			switch msg.String() {
			case "tab":
				if m.form.email.Focused() {
					m.form.email.Blur()
					m.form.pass.Focus()
				} else {
					m.form.pass.Blur()
					m.form.email.Focus()
				}
			case "enter", "space":
				m.form.email.Blur()
				m.form.pass.Blur()
				return m, registerCmd(m.client, m.form.email.Value(), m.form.pass.Value())
			}

		// -------------- обработка нажатия клавиш => Страница подтверждения почты
		case confirmPage:
			m.form.confirm.Blur()
			m.form.confirm.Focus()
			switch msg.String() {
			case "enter":
				return m, confirmEmailCmd(m.client, m.form.email.Value(), m.form.confirm.Value())
			}

		// -------------- обработка нажатия клавиш => Страница входа
		case loginPage:
			switch msg.String() {
			case "tab":
				if m.form.email.Focused() {
					m.form.email.Blur()
					m.form.pass.Focus()
				} else {
					m.form.pass.Blur()
					m.form.email.Focus()
				}
				return m, nil
			case "enter":
				return m, loginCmd(m.client, m.form.email.Value(), m.form.pass.Value())
			}

			// -------------- обработка нажатия клавиш => главная страница авторизации
		case homePage:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
				return m, nil
			case "enter", "space":
				m.form.email.SetValue("")
				m.form.pass.SetValue("")
				if m.choices[m.cursor] == "register" {
					m.page = registerPage
					m.history = append(m.history, registerPage)
				}
				if m.choices[m.cursor] == "login" {
					m.page = loginPage
					m.history = append(m.history, loginPage)
				}
				return m, nil
			}
		}
	}

	// --------------
	var cmd tea.Cmd
	if m.form.email.Focused() {
		m.form.email, cmd = m.form.email.Update(msg)
	} else if m.form.pass.Focused() {
		m.form.pass, cmd = m.form.pass.Update(msg)
	} else if m.form.confirm.Focused() {
		m.form.confirm, cmd = m.form.confirm.Update(msg)
	}
	return m, cmd
}
