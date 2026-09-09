package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

func dividerView(width int) string {
	return lipgloss.NewStyle().
		Foreground(colorMuted).Render(strings.Repeat("─", width))
}

func headerView(width int, left, right string) string {
	leftRendered := lipgloss.NewStyle().Foreground(colorPrimary).Render(left)
	rightRendered := right

	gapWidth := width - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func mainView(m model) string {
	s := ""
	if m.page == homePage {
		for i, choice := range m.choices {
			label := choice.String()
			if i == m.cursor {
				s += "* " + lipgloss.NewStyle().Foreground(colorText).Render(label)
			} else {
				s += "  " + lipgloss.NewStyle().Foreground(colorText).Render(label)
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
	return s
}

func footerView(width int, left, right string) string {
	leftRendered := lipgloss.NewStyle().Foreground(colorText).Faint(true).Render(left)
	rightRendered := lipgloss.NewStyle().Foreground(colorPrimary).Render(right)

	gapWidth := width - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func breadcrumbView(page, user string) string {
	text := page
	if user != "" {
		text += "/" + user
	}
	return text
}

func fieldView(label string, input textinput.Model) string {
	box := inputBoxBlurred
	if input.Focused() {
		box = inputBoxFocused
	}
	return labelStyle.Render(label) + "\n" + box.Render(input.View())
}
