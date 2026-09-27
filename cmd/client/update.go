package main

import (
	"errors"

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
	if m.selectedCard >= 0 {
		switch msg.String() {
		case "left":
			if m.activeCardBtn > 0 {
				m.activeCardBtn--
				return m, nil
			}
		case "right":
			if int(m.activeCardBtn) < len(m.cardsActions)-1 {
				m.activeCardBtn++
				return m, nil
			}
		case "enter":
			if m.activeCardBtn == cardActionsUpdate {
				m.isLoad = true
				return m, activateCardCmd(m.client, m.cards[m.selectedCard].GetId())
			}

			if m.activeCardBtn == cardActionsDelete {
				m.isLoad = true
				return m, deleteCardCmd(m.client, m.cards[m.selectedCard].GetId())
			}
		}
	} else {
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
		case "enter":
			m.profileForm.passOld.Focus()
			m.profileForm.passNew.Blur()
			m.profileForm.passOld.SetValue("")
			m.profileForm.passNew.SetValue("")
			return m.navigateTo(profileFormPage), nil
		}

	case tabCards:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.selectedCard = -1
			}
		case "down":
			if m.cursor < len(m.cards)-1 {
				m.cursor++
				m.selectedCard = -1
			}
		case "ctrl+a":
			m.cardForm.number.Focus()
			// Для теста
			m.cardForm.number.SetValue("5555640328235487")
			m.cardForm.date.SetValue("01/27")
			return m.navigateTo(cardFormPage), nil

		case "enter":
			if len(m.cards) > 0 {
				m.selectedCard = m.cursor
			}
		}

	case tabData:
		switch msg.String() {
		case "enter":
			return m.navigateTo(homeDirPage), nil
		}
	}
	return m, nil
}

func profileFormUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		var cmd tea.Cmd
		if m.profileForm.passOld.Focused() {
			m.profileForm.passOld.Blur()
			cmd = m.profileForm.passNew.Focus()
		} else {
			m.profileForm.passNew.Blur()
			cmd = m.profileForm.passOld.Focus()
		}
		return m, cmd
	case "enter":
		m.isLoad = true
		return m, changePassCmd(m.client, m.profileForm.passOld.Value(), m.profileForm.passNew.Value())

	}

	var cmd tea.Cmd
	switch {
	case m.profileForm.passOld.Focused():
		m.profileForm.passOld, cmd = m.profileForm.passOld.Update(msg)
	case m.profileForm.passNew.Focused():
		m.profileForm.passNew, cmd = m.profileForm.passNew.Update(msg)
	}
	return m, cmd
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
		return m, createCardCmd(m.client, m.cardForm.number.Value(), m.cardForm.date.Value())
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

func homeDirUpdate(msg tea.KeyPressMsg, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	if didSelect, _ := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.selectedFile = ""
		return m, m.handleError(errors.New("is not valid"))
	}

	switch msg.String() {
	case "ctrl+s":
		if m.selectedFile == "" {
			return m, m.handleError(errors.New("file is not selected"))
		}
		m.isLoad = false
		return m.navigateTo(userPage), nil
	}

	return m, cmd
}
