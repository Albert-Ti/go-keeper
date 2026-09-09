package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

func dividerView() string {
	return lipgloss.NewStyle().
		Foreground(colorMuted).Render(strings.Repeat("─", cardWidth))
}

func headerView(m model) string {
	spinner := ""
	if m.isLoad {
		spinner = m.spinner.View()
	}

	leftRendered := lipgloss.NewStyle().Foreground(colorPrimary).
		Render("Go Keeper") + " " + spinner

	rightRendered := breadcrumbView(m.activePage.String(), m.authUser)

	gapWidth := cardWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func mainView(m model) string {
	s := ""
	if m.activePage == homePage {
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

	if m.activePage == registerPage || m.activePage == loginPage {
		s += fieldView("Email", m.form.email) + "\n\n"
		s += fieldView("Password", m.form.pass) + "\n"
	}

	if m.activePage == confirmPage {
		s += "Keep a code to confirm your email: " + m.codeEmail + "\n\n"
		s += fieldView("Confirmation code", m.form.confirm) + "\n"
	}

	if m.activePage == userPage {
		var rendered []string
		for _, t := range m.allTabs {
			if t == m.activeTab {
				rendered = append(rendered, tabActiveStyle.Render(t.String()))
			} else {
				rendered = append(rendered, tabInactiveStyle.Render(t.String()))
			}
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
		s += "\n\n"
		switch m.activeTab {
		case tabProfile:
			s += "PROFILE\n"
		case tabCards:
			s += "CARDS\n"
		case tabData:
			s += "DATA\n"
		}

	}

	if m.textError != "" {
		s += errorStyle.Width(cardWidth).Align(lipgloss.Center).Render(m.textError) + "\n"
	} else {
		s += "\n"
	}
	return s
}

func footerView() string {
	leftRendered := lipgloss.NewStyle().Foreground(colorText).Faint(true).
		Render("ctrl+c quit· ctrl+q logout · esc back · tab focus")

	rightRendered := lipgloss.NewStyle().Foreground(colorPrimary).Render("© Albert Taygibov")

	gapWidth := cardWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func breadcrumbView(activePage, user string) string {
	text := activePage
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
