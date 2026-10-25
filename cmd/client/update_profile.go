package main

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) profileFormPageUpdate(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			m.historyBack()

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
