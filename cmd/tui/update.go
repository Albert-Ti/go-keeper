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

func userUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	// 1. Сначала обрабатываем фокус формы, если она активна
	if m.cardFormActive {
		switch msg.String() {
		case "tab":
			if m.cardForm.number.Focused() {
				m.cardForm.number.Blur()
				m.cardForm.date.Focus()
			} else {
				m.cardForm.date.Blur()
				m.cardForm.number.Focus()
			}
			return m, nil
		case "esc": // полезно добавить выход из формы
			m.cardForm.number.Blur()
			m.cardForm.date.Blur()
			m.cardFormActive = false
			return m, nil
		}
	}

	prevTab := m.activeTab

	switch msg.String() {
	case "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}
	case "right", "l":
		if int(m.activeTab) < len(m.allTabs)-1 {
			m.activeTab++
		}
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.contentTabProfile)-1 {
			m.cursor++
		}
	case "enter":
		m.cardFormActive = true
		m.cardForm.number.Focus()
		return m, nil
	}

	if m.activeTab == prevTab {
		return m, nil // вкладка не изменилась — ничего не запрашиваем
	}
	switch m.activeTab {
	case tabProfile:
		m.cursor = 0
		m.isLoad = true
		return m, getProfileCmd(m.client, m.accessToken)
	case tabCards:
		m.cursor = 0
		return m, getCardsCmd(m.client, m.accessToken)
	case tabData:
		m.cursor = 0
	}

	return m, nil
}
