package main

import tea "charm.land/bubbletea/v2"

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
