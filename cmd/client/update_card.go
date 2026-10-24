package main

import (
	tea "charm.land/bubbletea/v2"
)

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
