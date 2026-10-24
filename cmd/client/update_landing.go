package main

import (
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
