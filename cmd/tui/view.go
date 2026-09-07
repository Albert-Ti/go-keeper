package main

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func fieldView(label string, input textinput.Model) string {
	box := inputBoxBlurred
	if input.Focused() {
		box = inputBoxFocused
	}
	return labelStyle.Render(label) + "\n" + box.Render(input.View())
}

func (m model) View() tea.View {
	var content string
	s := headerView(cardWidth, "Go Keeper", breadcrumbView(m.page.String(), m.authUser)) + "\n"
	s += divider(cardWidth) + lipgloss.NewStyle().MarginBottom(2).Render("\n")

	if m.page == homePage {
		for i, choice := range m.choices {
			if i == m.cursor {
				s += "* " + lipgloss.NewStyle().Foreground(colorText).Render(choice)
			} else {
				s += "  " + lipgloss.NewStyle().Foreground(colorText).Render(choice)
			}
			s += "\n\n"
		}
	}
	if m.page == registerPage || m.page == loginPage {
		s += fieldView("Email", m.form.email) + "\n\n"
		s += fieldView("Password", m.form.pass) + "\n"
	}
	if m.page == confirmPage {
		s += "Keep a code to confirm your email: " + m.codeEmail + "\n\n"
		s += fieldView("Confirmation code", m.form.confirm) + "\n"
	}
	if m.page == profilePage {
		s += "PROFILE USER \n\n"
	}
	if m.textError != "" {
		s += errorStyle.Width(cardWidth).Align(lipgloss.Center).Render(m.textError) + "\n"
	} else {
		s += "\n"
	}
	s += divider(cardWidth) + "\n"
	s += footerView(cardWidth, "ctrl+c quit · esc discard · tab focus", "© Albert Taygibov")
	content = s

	centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)

	// фон для отладки
	// debugBackground := lipgloss.NewStyle().
	// 	Width(m.width).
	// 	Height(m.height).
	// 	Background(lipgloss.Color("#1a1a2e")).
	// 	Foreground(lipgloss.Color("#e0e0e0"))
	// return tea.NewView(debugBackground.Render(centered))
	return tea.NewView(centered)
}
