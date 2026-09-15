package main

import (
	tea "charm.land/bubbletea/v2"
)

func homeUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
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
		return m.navigateTo(m.choices[m.cursor]), nil
	}
	return m, nil
}

func registerUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.authForm.email.Focused() {
			m.authForm.email.Blur()
			m.authForm.pass.Focus()
		} else {
			m.authForm.pass.Blur()
			m.authForm.email.Focus()
		}
		return m, nil
	case "enter", "space":
		m.isLoad = true
		return m, registerCmd(m.client, m.authForm.email.Value(), m.authForm.pass.Value())
	}

	var cmd tea.Cmd
	if m.authForm.email.Focused() {
		m.authForm.email, cmd = m.authForm.email.Update(msg)
	} else if m.authForm.pass.Focused() {
		m.authForm.pass, cmd = m.authForm.pass.Update(msg)
	}
	return m, cmd
}

func confirmUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	m.authForm.pass.Blur()
	m.authForm.email.Blur()
	m.authForm.confirm.Focus()

	switch msg.String() {
	case "enter":
		m.isLoad = true
		return m, confirmEmailCmd(m.client, m.authForm.email.Value(), m.authForm.confirm.Value())
	}

	var cmd tea.Cmd
	m.authForm.confirm, cmd = m.authForm.confirm.Update(msg)
	return m, cmd
}

func loginUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.authForm.email.Focused() {
			m.authForm.email.Blur()
			m.authForm.pass.Focus()
		} else {
			m.authForm.pass.Blur()
			m.authForm.email.Focus()
		}
		return m, nil
	case "enter":
		m.isLoad = true
		return m, loginCmd(m.client, m.authForm.email.Value(), m.authForm.pass.Value())
	}

	var cmd tea.Cmd
	if m.authForm.email.Focused() {
		m.authForm.email, cmd = m.authForm.email.Update(msg)
	} else if m.authForm.pass.Focused() {
		m.authForm.pass, cmd = m.authForm.pass.Update(msg)
	}
	return m, cmd
}

func loadActiveTab(activeTab tabType, m model) tea.Cmd {
	switch activeTab {
	case tabProfile:
		return getProfileCmd(m.client)
	case tabCards:
		return getCardsCmd(m.client)
	case tabData:
	}
	return nil
}

func userUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	switch msg.String() {
	case "left":
		if m.activeTab > 0 {
			m.activeTab--
			m.cursor = 0
			m.isLoad = true
			return m, loadActiveTab(m.activeTab, m)
		}
	case "right":
		if int(m.activeTab) < len(m.allTabs)-1 {
			m.activeTab++
			m.cursor = 0
			m.isLoad = true
			return m, loadActiveTab(m.activeTab, m)
		}
	}

	switch m.activeTab {
	case tabProfile:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.contentTabProfile)-1 {
				m.cursor++
			}
		}

	case tabCards:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.cards)-1 {
				m.cursor++
			}
		case "enter":
			m = m.navigateTo(cardFormPage)
			cmd := m.cardForm.number.Focus()
			return m, cmd
		}

	case tabData:
	}
	return m, nil
}

func cardFormUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		var cmd tea.Cmd
		if m.cardForm.number.Focused() {
			m.cardForm.number.Blur()
			cmd = m.cardForm.date.Focus()
		} else {
			m.cardForm.date.Blur()
			cmd = m.cardForm.number.Focus()
		}
		return m, cmd
	case "enter":
		m.isLoad = true
		cmd := createCardCmd(m.client, m.accessToken,
			m.cardForm.number.Value())
		return m, cmd
	}

	var cmd tea.Cmd
	switch {
	case m.cardForm.number.Focused():
		m.cardForm.number, cmd = m.cardForm.number.Update(msg)
	case m.cardForm.date.Focused():
		m.cardForm.date, cmd = m.cardForm.date.Update(msg)
	}
	return m, cmd
}
