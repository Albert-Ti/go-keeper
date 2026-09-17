package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

func dividerView() string {
	return lipgloss.NewStyle().Faint(true).Render(strings.Repeat("─", cardWidth))
}

func headerView(m model) string {
	spinner := ""
	if m.isLoad {
		spinner = m.spinner.View()
	}

	leftRendered := lipgloss.NewStyle().Foreground(colorPrimary).
		Render("Go Keeper") + " " + spinner

	rightRendered := m.activePage.String()

	gapWidth := cardWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func mainView(m model) string {
	s := ""

	if len(m.history) > 1 {
		s += lipgloss.NewStyle().
			Width(cardWidth).     // Устанавливаем ширину как у карточки
			Align(lipgloss.Left). // Прижимаем к левому краю
			Faint(true).
			Render("← "+m.history[len(m.history)-2].String()) + "\n\n"
	} else {
		s += "\n\n"
	}

	if m.activePage == homePage {
		s += "\n\n"
		s += "Welcome to the Go Keeper project! version 1.0.0\n\n\n"
		for i, choice := range m.choices {
			label := choice.String()
			if i == m.cursor {
				s += "> " + lipgloss.NewStyle().Render(strings.ToUpper(label))
			} else {
				s += "  " + lipgloss.NewStyle().Render(strings.ToUpper(label))
			}
			s += "\n\n"
		}
	}

	if m.activePage == registerPage || m.activePage == loginPage {
		s += fieldView("Email", m.authForm.email) + "\n\n"
		s += fieldView("Pass", m.authForm.pass) + "\n"
	}

	if m.activePage == confirmPage {
		s += "Keep a code to confirm your email: " + m.codeEmail + "\n\n"
		s += fieldView("Confirmation code", m.authForm.confirm) + "\n"
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
			s += contentProfileView(m)
		case tabCards:
			s += contentCardsView(m)
		case tabData:
			s += "DATA\n"
		}
	}

	if m.activePage == cardFormPage {
		s += "\n\n" + fieldView("Card number", m.cardForm.number) + "\n"
		s += fieldView("Expiry date", m.cardForm.date) + "\n"
	}

	if m.activePage == profileFormPage {
		s += "\n\n" + fieldView("Old pass", m.profileForm.passOld) + "\n"
		s += fieldView("New pass", m.profileForm.passNew) + "\n"
	}

	if m.textError != "" {
		s += "\n" + errorStyle.Width(cardWidth).Align(lipgloss.Center).Render(m.textError) + "\n"
	}

	// Добавляем пустые строки, чтобы заполнить пространство
	currentHeight := lipgloss.Height(s)
	if currentHeight < minHeight {
		s += strings.Repeat("\n", minHeight-currentHeight)
	}

	return s
}

func footerView() string {
	leftRendered := lipgloss.NewStyle().Faint(true).
		Render("ctrl+c quit· ctrl+q logout · esc back · tab focus")

	rightRendered := lipgloss.NewStyle().Foreground(colorPrimary).Render("© Albert Taygibov")

	gapWidth := cardWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func fieldView(label string, input textinput.Model) string {
	box := inputBoxBlurred
	if input.Focused() {
		box = inputBoxFocused
	}
	return labelStyle.Render(label) + "\n" + box.Render(input.View())
}

func contentProfileView(m model) string {
	s := ""
	for i, v := range m.contentTabProfile {
		if i == m.cursor {
			s += "✎ " + lipgloss.NewStyle().Width(12).Render(v) +
				": " + m.profile[v] + "\n"
		} else {
			s += "  " + lipgloss.NewStyle().Width(12).Render(v) +
				": " + m.profile[v] + "\n"
		}
	}
	s += "\n" + "  " + buttonEnter + lipgloss.NewStyle().Faint(true).Render(" press to update info") + "\n"

	return lipgloss.NewStyle().MarginTop(1).PaddingLeft(3).Width(cardWidth).Align(lipgloss.Left).Render(s)
}

func contentCardsView(m model) string {
	s := ""
	if len(m.cards) == 0 {
		textInfo := lipgloss.NewStyle().Faint(true).Render("You have not added a bank card yet, to add : ")
		s = "\n\n" + textInfo + buttonEnter + "\n"
	} else {
		for i, card := range m.cards {
			isActive := lipgloss.NewStyle().Faint(true).Render("inactive")
			if card.GetActive() {
				isActive = "active"
			}
			prefix := "  "
			if i == m.cursor {
				prefix = "✎ "
			}
			s += prefix + lipgloss.NewStyle().Faint(true).Render("**** **** **** ") + card.GetCardNumber() + "  " + card.GetExpiryDate().AsTime().Format("02/06") + "  " + isActive + "\n"
		}
		textInfo := lipgloss.NewStyle().Faint(true).Render(" press to add new card.")
		s += "\n" + "  " + buttonEnter + textInfo + "\n"
	}

	return lipgloss.NewStyle().MarginTop(1).PaddingLeft(3).Width(cardWidth).Align(lipgloss.Left).Render(s)
}
