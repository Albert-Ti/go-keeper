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
		if m.form.email.Focused() {
			m.form.email.Blur()
			m.form.pass.Focus()
		} else {
			m.form.pass.Blur()
			m.form.email.Focus()
		}
		return m, nil
	case "enter", "space":
		m.isLoad = true
		return m, registerCmd(m.client, m.form.email.Value(), m.form.pass.Value())
	}

	var cmd tea.Cmd
	if m.form.email.Focused() {
		m.form.email, cmd = m.form.email.Update(msg)
	} else if m.form.pass.Focused() {
		m.form.pass, cmd = m.form.pass.Update(msg)
	}
	return m, cmd
}

func confirmUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	m.form.pass.Blur()
	m.form.email.Blur()
	m.form.confirm.Focus()
	m.form.confirm.SetValue("")

	switch msg.String() {
	case "enter":
		m.isLoad = true
		return m, confirmEmailCmd(m.client, m.form.email.Value(), m.form.confirm.Value())
	}

	var cmd tea.Cmd
	m.form.confirm, cmd = m.form.confirm.Update(msg)
	return m, cmd
}

func loginUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
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
		m.isLoad = true
		return m, loginCmd(m.client, m.form.email.Value(), m.form.pass.Value())
	}

	var cmd tea.Cmd
	if m.form.email.Focused() {
		m.form.email, cmd = m.form.email.Update(msg)
	} else if m.form.pass.Focused() {
		m.form.pass, cmd = m.form.pass.Update(msg)
	}
	return m, cmd
}

func userUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}
	case "right", "l":
		if int(m.activeTab) < len(m.allTabs)-1 {
			m.activeTab++
		}
	}
	switch m.activeTab {
	case tabProfile:
		cmd := getProfileCmd(m.client, m.accessToken)
		switch msg.String() {
		case "up":
			if len(m.user) > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.user)-1 {
				m.cursor++
			}
		case "enter":
		}
		m.isLoad = true
		return m, cmd
	case tabCards:
		return m, getCardsCmd(m.client, m.accessToken)
	case tabData:
	}

	return m, nil
}
