package main

import (
	"errors"

	tea "charm.land/bubbletea/v2"
)

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

		if !m.privateData {
			return m, m.handleInfo("you need used master key", "filePickerPageUpdate")
		}
		m.isLoad = true
		return m, tea.Batch(
			createArbitraryDataCmd(
				m.client,
				m.filepicker.Path,
				m.localStorage.Get(masterKeyKey),
				m.progressCh,
			),
			waitUploadProgress(m.progressCh),
		)
	}

	return m, cmd
}
