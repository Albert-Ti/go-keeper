package main

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/dustin/go-humanize"
)

func headerView(m model) string {
	spinner := ""
	if m.isLoad {
		spinner = m.spinner.View()
	}

	leftRendered := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Render("Go Keeper") + " " + spinner

	page := m.activePage.String()
	if m.activePage >= homePage {
		page = m.authUser
	}
	rightRendered := page

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

	if m.activePage == landingPage {
		s += "\n"
		s += "Welcome to the Go Keeper project!\n\n\n\n"
		for i, choice := range m.choices {
			if i == m.cursor {
				s += "> " + lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(choice.String()) + " <"
			} else {
				s += "  " + lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(choice.String()) + "  "
			}
			s += "\n\n"
		}
	}

	if m.activePage == registerPage || m.activePage == loginPage {
		switch m.activePage {
		case registerPage:
			s += lipgloss.NewStyle().Foreground(colorAccent).Render("REGISTRATION") + " \n\n"
		case loginPage:
			s += lipgloss.NewStyle().Foreground(colorAccent).Render("LOGIN") + " \n\n"
		}
		s += fieldView("Email", m.authForm.email) + "\n\n"
		s += fieldView("Pass", m.authForm.pass) + "\n"
	}

	if m.activePage == confirmPage {
		s += lipgloss.NewStyle().Foreground(colorAccent).Render("CONFIRM EMAIL") + " \n\n"
		s += "Keep a code to confirm your email: " + m.codeEmail + "\n\n"
		s += fieldView("Confirmation code", m.authForm.confirm) + "\n"
	}

	if m.activePage == homePage {
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
			s += contentDataView(m)
		case tabLocalData:
			s += contentLocalDataView(m)
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

	if m.activePage == filePickerPage {
		if m.selectedFile == "" {
			s += "Pick a file:"
		} else {
			s += "Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile)
		}
		s += "\n\n" + m.filepicker.View() + "\n"
		disabled := m.selectedFile == ""
		textBtn := ""
		if !disabled {
			textBtn = "ctrl+s submit to server"
		}
		s += "\n" + button("submit", textBtn, disabled) + "\n\n"
	}

	if m.textError != "" {
		s += "\n\n" + errorStyle.Width(cardWidth).Align(lipgloss.Center).Render(m.textError) + "\n"
	}
	if m.textInfo != "" {
		s += "\n\n" + infoStyle.Width(cardWidth).Align(lipgloss.Center).Render(m.textInfo) + "\n"
	}

	// Добавляем пустые строки, чтобы заполнить пространство
	currentHeight := lipgloss.Height(s)
	if currentHeight < minHeight {
		s += strings.Repeat("\n", minHeight-currentHeight)
	}
	s += lipgloss.NewStyle().Faint(true).Width(cardWidth).Align(lipgloss.Left).
		Render("ctrl+c quit· ctrl+q logout · esc back · tab focus") + "\n"
	return s
}

func footerView() string {
	leftRendered := lipgloss.NewStyle().Faint(true).
		Render("build " + VERSION)

	rightRendered := lipgloss.NewStyle().Foreground(colorAccent).Render("© Albert Taygibov")

	gapWidth := cardWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}

func dividerView() string {
	return lipgloss.NewStyle().Faint(true).Render(strings.Repeat("─", cardWidth))
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
	for _, v := range m.contentTabProfile {
		s += "  " + lipgloss.NewStyle().Width(12).Render(v) +
			": " + m.profile[v] + "\n"
	}
	s += "  " + button("update", "press enter to update info", false) + "\n"

	return lipgloss.NewStyle().MarginTop(1).PaddingLeft(3).Width(cardWidth).Align(lipgloss.Left).Render(s)
}

func contentCardsView(m model) string {
	s := ""
	if len(m.cards) == 0 {
		textInfo := lipgloss.NewStyle().Faint(true).Render("You have not added a bank card yet, press ctrl+a to ")
		s = textInfo + button("add", "", false) + "\n"

		return lipgloss.NewStyle().Width(cardWidth).Align(lipgloss.Center).Render(s)
	} else {
		for i, card := range m.cards {
			isActive := lipgloss.NewStyle().Faint(true).Render("inactive")

			if card.GetActive() {
				isActive = "active  "
			}

			cardsBtns := []string{}
			prefix := "  "
			if i == m.cursor {
				prefix = "✎ "

				if m.selectedRowID == i {
					cardsBtns = []string{"[ ]", "[ delete ]"}
					if card.GetActive() {
						cardsBtns = []string{"[X]", "[ delete ]"}
					}
				}
			}
			card := prefix + strconv.Itoa(i+1) + ". " +
				lipgloss.NewStyle().Faint(true).Render("**** **** **** ") + card.GetCardNumber() + "  " + card.GetExpiryDate().AsTime().Format("01/06") + "  " + isActive

			btns := ""
			for i, v := range cardsBtns {
				active := lipgloss.NewStyle().Faint(true).Foreground(colorAccent).Render(v)
				if i == m.activeBtn {
					active = lipgloss.NewStyle().Foreground(colorAccent).Render(v)
				}
				btns += active + " "
			}

			s += lipgloss.NewStyle().MarginRight(5).Render(card) + btns + "\n"
		}
		s += "\n" + "  " + button("add", "ctrl+a", false) + "\n"
		s += "  " + button("select", "enter", false) + "\n"
	}

	return lipgloss.NewStyle().MarginTop(1).PaddingLeft(3).Width(cardWidth).Align(lipgloss.Left).Render(s)
}

func contentDataView(m model) string {
	var (
		nameWidth   = 28
		typeWidth   = 8
		sizeWidth   = 19
		statusWidth = 10
		dateWidth   = 13
	)

	s := ""

	if len(m.arbitraryData) > 0 {
		s += "  " +
			lipgloss.NewStyle().Width(nameWidth).Bold(true).Render("name") +
			lipgloss.NewStyle().Width(typeWidth).Bold(true).Render("type") +
			lipgloss.NewStyle().Width(sizeWidth).Bold(true).Render("size") +
			lipgloss.NewStyle().Width(dateWidth).Bold(true).Render("date") +
			lipgloss.NewStyle().Width(statusWidth).Bold(true).Render("status") + "\n\n"

		// Строки
		for i, v := range m.arbitraryData {
			filename := v.GetName()
			if len(filename) > 25 {
				filename = v.GetName()[:20] + "..."
			}

			var status string

			var size = humanize.Bytes(uint64(v.GetClientSize())) + " → " + humanize.Bytes(uint64(v.GetTotalSize()))
			switch v.GetStatus() {
			case 0:
				status = lipgloss.NewStyle().Faint(true).Render("uploading")
			case 1:
				status = lipgloss.NewStyle().Foreground(colorSuccess).Render("success")
			case 2:
				status = lipgloss.NewStyle().Foreground(colorError).Render("failed")
			}

			dataBtns := []string{}
			prefix := "  "
			if i == m.cursor {
				prefix = "✎ "

				if m.selectedRowID == i {
					dataBtns = []string{"[↓]", "[reload]", "[delete]"}
				}
			}
			btns := ""
			for i, v := range dataBtns {
				active := lipgloss.NewStyle().Faint(true).Foreground(colorAccent).Render(v)
				if i == m.activeBtn {
					active = lipgloss.NewStyle().Foreground(colorAccent).Render(v)
				}
				btns += active + " "
			}

			btnsOrStatus := ""
			if btns == "" {
				btnsOrStatus = lipgloss.NewStyle().Width(dateWidth).
					Render(v.GetCreatedAt().AsTime().Format("02.01.2006")) +
					lipgloss.NewStyle().Width(statusWidth).Render(status)
			} else {

				btnsOrStatus = btns
			}

			s += prefix +
				lipgloss.NewStyle().Width(nameWidth).Foreground(colorAccent).Render(strconv.Itoa(i+1)+". "+filename) +
				lipgloss.NewStyle().Width(typeWidth).Render(v.GetType()) +
				lipgloss.NewStyle().Width(sizeWidth).Render(size) +
				btnsOrStatus + "\n"
		}
		s += "\n"
		s += "  " + button("add", "ctrl+a", false) + "\n"
		s += "  " + button("select", "enter", false)

		return lipgloss.NewStyle().Width(cardWidth).Align(lipgloss.Left).MarginTop(1).Render(s)
	}

	textInfo := lipgloss.NewStyle().Faint(true).Render("You have not added a files yet, press ctrl+a to ")
	s = "\n\n" + textInfo + button("add", "", false) + "\n"

	return lipgloss.NewStyle().PaddingLeft(3).Align(lipgloss.Center).Render(s)
}

func contentLocalDataView(m model) string {
	s := ""
	sData := ""
	title := "Local list of uploaded data:"

	if len(m.localArbitraryData) > 0 {
		for _, v := range m.localArbitraryData {
			sData += " | " + v.filename + "  " + humanize.Bytes(uint64(v.size)) + "\n"
		}
	} else {
		sData += "Empty"
	}

	s += "\n" + lipgloss.NewStyle().Width(cardWidth).Align(lipgloss.Left).Foreground(colorAccent).Render(title) + "\n\n"

	s += lipgloss.NewStyle().Width(cardWidth).Align(lipgloss.Left).Foreground(colorMuted).Render(sData)

	return s
}
