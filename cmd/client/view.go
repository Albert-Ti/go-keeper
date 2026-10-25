package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	var content string
	s := m.headerView() + "\n"
	s += dividerView() + lipgloss.NewStyle().Render("\n")

	s += m.mainView()

	s += dividerView() + "\n"
	s += footerView()

	content = s
	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	v := tea.NewView(centered)
	v.AltScreen = true
	return v
}

func dividerView() string {
	return lipgloss.NewStyle().Faint(true).Render(strings.Repeat("─", maxWidth))
}

func fieldView(label string, input textinput.Model) string {
	box := inputBoxBlurred
	if input.Focused() {
		box = inputBoxFocused
	}
	return labelStyle.Render(label) + "\n" + box.Render(input.View())
}

func (m model) progressView() string {
	percent := float64(m.uploadedBytes) / float64(m.totalBytes)

	s := ""
	pad := strings.Repeat(" ", padding)
	return s + "\n" +

		pad + m.progress.ViewAs(percent) + "\n\n"
}
