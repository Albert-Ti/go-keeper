package main

import (
	"errors"

	tea "charm.land/bubbletea/v2"
)

func (m model) landingPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
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
	case "enter":
		cmd := m.authForm.email.Focus()
		m.authForm.pass.Blur()
		return m.navigateTo(m.choices[m.cursor]), cmd
	}
	return m, nil
}

func (m model) registerPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.authForm.email.Focused() {
			m.authForm.email.Blur()
			return m, m.authForm.pass.Focus()
		} else {
			m.authForm.pass.Blur()
			return m, m.authForm.email.Focus()
		}
	case "enter":
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

func (m model) confirmPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.isLoad = true
		return m, tea.Batch(
			confirmEmailCmd(m.client, m.authForm.email.Value(), m.authForm.confirm.Value()),
		)
	}

	var cmd tea.Cmd
	m.authForm.confirm, cmd = m.authForm.confirm.Update(msg)
	return m, cmd
}

func (m model) loginPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.authForm.email.Focused() {
			m.authForm.email.Blur()
			return m, m.authForm.pass.Focus()
		} else {
			m.authForm.pass.Blur()
			return m, m.authForm.email.Focus()
		}

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

func (m model) loadActiveTab(activeTab tabType) tea.Cmd {
	switch activeTab {
	case tabProfile:
		m.isLoad = true
		return getProfileCmd(m.client)
	case tabCards:
		m.isLoad = true
		return getCardsCmd(m.client)
	case tabData:
		m.isLoad = true
		return getArbitraryDataCmd(m.client)
	case tabLocalData:
		m.isLoad = true
		return getLocalArbitraryDataCmd()
	}
	return nil
}

func (m model) homePageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
	if m.selectedRowID >= 0 {
		switch m.activeTab {
		case tabCards:
			switch msg.String() {
			case "left":
				if m.activeBtn > 0 {
					m.activeBtn--
					return m, nil
				}
			case "right":
				if m.activeBtn < len(m.cardsActions)-1 {
					m.activeBtn++
					return m, nil
				}
			case "enter":
				if m.activeBtn == int(cardActionsUpdate) {
					m.isLoad = true
					if m.cards[m.selectedRowID].GetActive() {
						return m, m.handleInfo("card is already active", "")
					}
					return m, activateCardCmd(m.client, m.cards[m.selectedRowID].GetId())
				}

				if m.activeBtn == int(cardActionsDelete) {
					m.isLoad = true
					return m, deleteCardCmd(m.client, m.cards[m.selectedRowID].GetId())
				}
			}
		case tabData:
			switch msg.String() {
			case "left":
				if m.activeBtn > 0 {
					m.activeBtn--
					return m, nil
				}
			case "right":
				if m.activeBtn < len(m.dataActions)-1 {
					m.activeBtn++
					return m, nil
				}
			case "enter":
				if m.activeBtn == int(dataActionsDownload) {
					m.isLoad = true
					userLocalPath := DOWNLOAD_PATHNAME + "/" + m.profile["email"]

					return m, tea.Batch(
						downloadArbitraryDataCmd(
							m.client,
							m.arbitraryData[m.selectedRowID].GetId(),
							m.arbitraryData[m.selectedRowID].GetName(),
							userLocalPath,
							m.progressCh,
						),
						waitUploadProgress(m.progressCh),
					)
				}

				if m.activeBtn == int(dataActionsReload) {
					m.isLoad = true
					if m.arbitraryData[m.selectedRowID].GetStatus() == 1 {
						return m, m.handleInfo("file has been successfully uploaded", "tabData")
					}
					id := m.arbitraryData[m.selectedRowID].GetId()
					filepath := m.arbitraryData[m.selectedRowID].GetOsPath()

					return m, tea.Batch(
						reloadArbitraryDataCmd(m.client, id, filepath, m.progressCh),
						waitUploadProgress(m.progressCh),
					)
				}

				if m.activeBtn == int(dataActionsDelete) {
					m.isLoad = true
					return m, deleteArbitraryDataCmd(m.client, m.arbitraryData[m.selectedRowID].GetId())
				}
			}
		}
	} else {
		m.authForm.pass.Blur()
		m.authForm.email.Blur()
		m.authForm.confirm.Blur()
		switch msg.String() {
		case "left":
			if m.activeTab > 0 {
				m.activeTab--
				m.cursor = 0
				m.activeBtn = 0
				return m, m.loadActiveTab(m.activeTab)
			}
		case "right":
			if int(m.activeTab) < len(m.allTabs)-1 {
				m.activeTab++
				m.cursor = 0
				m.activeBtn = 0
				return m, m.loadActiveTab(m.activeTab)
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
			cmd := m.profileForm.passCurr.Focus()
			m.profileForm.passNew.Blur()
			m.profileForm.passCurr.SetValue("")
			m.profileForm.passNew.SetValue("")
			return m.navigateTo(profileFormPage), cmd
		}

	case tabCards:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.selectedRowID = -1
			}
		case "down":
			if m.cursor < len(m.cards)-1 {
				m.cursor++
				m.selectedRowID = -1
			}
		case "enter":
			if len(m.cards) > 0 {
				m.selectedRowID = m.cursor
			}
		case "ctrl+a":
			cmd := m.cardForm.number.Focus()
			return m.navigateTo(cardFormPage), cmd
		}

	case tabData:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.selectedRowID = -1
			}
		case "down":
			if m.cursor < len(m.arbitraryData)-1 {
				m.cursor++
				m.selectedRowID = -1
			}
		case "enter":
			if len(m.arbitraryData) > 0 {
				m.selectedRowID = m.cursor
			}

		case "ctrl+a":
			return m.navigateTo(filePickerPage), nil
		}

	case tabLocalData:
		switch msg.String() {
		case "enter":
		}
	}
	return m, nil
}

func (m model) profileFormPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		var cmd tea.Cmd
		if m.profileForm.passCurr.Focused() {
			m.profileForm.passCurr.Blur()
			cmd = m.profileForm.passNew.Focus()
		} else {
			m.profileForm.passNew.Blur()
			cmd = m.profileForm.passCurr.Focus()
		}
		return m, cmd
	case "enter":
		m.isLoad = true
		return m, changePassCmd(m.client, m.profileForm.passCurr.Value(), m.profileForm.passNew.Value())

	}

	var cmd tea.Cmd
	switch {
	case m.profileForm.passCurr.Focused():
		m.profileForm.passCurr, cmd = m.profileForm.passCurr.Update(msg)
	case m.profileForm.passNew.Focused():
		m.profileForm.passNew, cmd = m.profileForm.passNew.Update(msg)
	}
	return m, cmd
}

func (m model) cardFormPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
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

func (m model) filePickerPageUpdate(msg tea.KeyPressMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	if didSelect, _ := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.selectedFile = ""
		return m, m.handleError(errors.New("is not valid"), "filePickerPageUpdate")
	}

	switch msg.String() {
	case "ctrl+s":
		if m.selectedFile == "" {
			return m, m.handleError(errors.New("file is not selected"), "filePickerPageUpdate")
		}
		m.isLoad = true
		return m, tea.Batch(
			createArbitraryDataCmd(
				m.client,
				m.filepicker.Path,
				m.progressCh,
			),
			waitUploadProgress(m.progressCh),
		)
	case "ctrl+p":
		if m.privateData {
			m.privateData = false
			return m, cmd

		}
		m.privateData = true
		return m, cmd
	}

	return m, cmd
}

func (m model) updateNonKeyComponents(msg tea.Msg) (model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	m.filepicker, cmd = m.filepicker.Update(msg)
	cmds = append(cmds, cmd)

	switch {
	case m.authForm.email.Focused():
		m.authForm.email, cmd = m.authForm.email.Update(msg)
		cmds = append(cmds, cmd)

	case m.authForm.pass.Focused():
		m.authForm.pass, cmd = m.authForm.pass.Update(msg)
		cmds = append(cmds, cmd)

	case m.authForm.confirm.Focused():
		m.authForm.confirm, cmd = m.authForm.confirm.Update(msg)
		cmds = append(cmds, cmd)

	case m.cardForm.number.Focused():
		m.cardForm.number, cmd = m.cardForm.number.Update(msg)
		cmds = append(cmds, cmd)

	case m.cardForm.date.Focused():
		m.cardForm.date, cmd = m.cardForm.date.Update(msg)
		cmds = append(cmds, cmd)

	case m.profileForm.passCurr.Focused():
		m.profileForm.passCurr, cmd = m.profileForm.passCurr.Update(msg)
		cmds = append(cmds, cmd)

	case m.profileForm.passNew.Focused():
		m.profileForm.passNew, cmd = m.profileForm.passNew.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
