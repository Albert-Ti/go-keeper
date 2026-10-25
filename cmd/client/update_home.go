package main

import (
	"crypto/aes"

	tea "charm.land/bubbletea/v2"
)

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
		return getArbitraryDataCmd(m.client, m.localStorage.Get(masterKeyKey))
	case tabLocalData:
		m.isLoad = true
		return getLocalArbitraryDataCmd()
	}
	return nil
}

func (m model) homePageUpdate(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.selectedRowID >= 0 {
				m.selectedRowID = -1
				return m, nil
			}
			if len(m.history) <= 1 {
				return m, m.handleInfo("to log out, press ctrl+q", "")
			}
			m.historyBack()

		case "left", "right":
			if m.selectedRowID < 0 {
				return m.updateTabNavigation(msg)
			}
			return m.updateActionNavigation(msg)

		case "up", "down":
			if m.selectedRowID >= 0 {
				return m, nil
			}
			return m.updateRowNavigation(msg)

		case "enter":
			if m.selectedRowID >= 0 {
				return m.executeSelectedAction()
			}
			return m.selectCurrentRow()

		case "ctrl+a":
			return m.handleCtrlA()
		}

	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	return m, cmd
}

func (m model) updateTabNavigation(msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "left":
		if m.activeTab == 0 {
			return m, nil
		}

		m.activeTab--
		m.cursor = 0
		m.activeBtn = 0

	case "right":
		if int(m.activeTab) >= len(m.allTabs)-1 {
			return m, nil
		}

		m.activeTab++
		m.cursor = 0
		m.activeBtn = 0
	}

	return m, m.loadActiveTab(m.activeTab)
}

func (m model) updateActionNavigation(msg tea.KeyPressMsg) (model, tea.Cmd) {
	var actionsLen int

	switch m.activeTab {
	case tabCards:
		actionsLen = len(m.cardsActions)

	case tabData:
		actionsLen = len(m.dataActions)

	default:
		return m, nil
	}

	switch msg.String() {
	case "left":
		if m.activeBtn > 0 {
			m.activeBtn--
		}

	case "right":
		if m.activeBtn < actionsLen-1 {
			m.activeBtn++
		}
	}

	return m, nil
}

func (m model) updateRowNavigation(msg tea.KeyPressMsg) (model, tea.Cmd) {
	var rows int

	switch m.activeTab {
	case tabCards:
		rows = len(m.cards)

	case tabData:
		rows = len(m.arbitraryData)

	default:
		return m, nil
	}

	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
			m.selectedRowID = -1
		}

	case "down":
		if m.cursor < rows-1 {
			m.cursor++
			m.selectedRowID = -1
		}
	}

	return m, nil
}

func (m model) executeSelectedAction() (model, tea.Cmd) {
	switch m.activeTab {
	case tabCards:
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
		return m, nil

	case tabData:
		if m.activeBtn == int(dataActionsDownload) {
			m.isLoad = true
			userLocalPath := DOWNLOAD_PATHNAME + "/" + m.profile.email

			return m, tea.Batch(
				downloadArbitraryDataCmd(
					m.client,
					m.arbitraryData[m.selectedRowID].id,
					m.arbitraryData[m.selectedRowID].filename,
					m.arbitraryData[m.selectedRowID].filetype,
					userLocalPath,
					m.progressCh,
				),
				waitUploadProgress(m.progressCh),
			)
		}

		if m.activeBtn == int(dataActionsReload) {
			m.isLoad = true
			if m.arbitraryData[m.selectedRowID].status == 1 {
				return m, m.handleInfo("file has been successfully uploaded", "tabData")
			}
			id := m.arbitraryData[m.selectedRowID].id
			filepath := m.arbitraryData[m.selectedRowID].osPath

			return m, tea.Batch(
				reloadArbitraryDataCmd(m.client, id, filepath, m.progressCh),
				waitUploadProgress(m.progressCh),
			)
		}

		if m.activeBtn == int(dataActionsDelete) {
			m.isLoad = true
			return m, deleteArbitraryDataCmd(m.client, m.arbitraryData[m.selectedRowID].id)
		}

		return m, nil
	default:
		return m, nil
	}
}

func (m model) selectCurrentRow() (model, tea.Cmd) {
	switch m.activeTab {
	case tabProfile:
		cmd := m.profileForm.passCurr.Focus()
		m.profileForm.passNew.Blur()
		m.profileForm.passCurr.SetValue("")
		m.profileForm.passNew.SetValue("")
		return m.navigateTo(profileFormPage), cmd
	case tabCards:
		if len(m.cards) > 0 {
			m.selectedRowID = m.cursor
		}
		return m, nil
	case tabData:
		if len(m.arbitraryData) > 0 {
			m.selectedRowID = m.cursor
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m model) handleCtrlA() (model, tea.Cmd) {
	switch m.activeTab {
	case tabProfile:
		if m.localStorage.Get(masterKeyKey) == "" {
			masterKeyValue, err := generateRandom(aes.BlockSize)

			if err != nil {
				return m, m.handleError(err, "generateRandom")
			}

			if err := m.localStorage.Set(masterKeyKey, string(masterKeyValue)); err != nil {
				return m, m.handleError(err, "localStorageSet")
			}
			return m, nil
		}
		return m, m.handleInfo("you already have a master pass", "")
	case tabCards:
		cmd := m.cardForm.number.Focus()
		return m.navigateTo(cardFormPage), cmd
	case tabData:
		return m.navigateTo(filePickerPage), nil
	default:
		return m, nil
	}
}
