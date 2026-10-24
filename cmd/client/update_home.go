package main

import (
	"crypto/aes"

	tea "charm.land/bubbletea/v2"
)

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
		case "enter":
			cmd := m.profileForm.passCurr.Focus()
			m.profileForm.passNew.Blur()
			m.profileForm.passCurr.SetValue("")
			m.profileForm.passNew.SetValue("")
			return m.navigateTo(profileFormPage), cmd

		case "ctrl+a":
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
	}
	return m, nil
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
		return getArbitraryDataCmd(m.client, m.localStorage.Get(masterKeyKey))
	case tabLocalData:
		m.isLoad = true
		return getLocalArbitraryDataCmd()
	}
	return nil
}
