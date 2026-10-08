package main

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

const cardWidth = 80

// Устанавливаем минимальную высоту для контента
const minHeight = 18

// Color
var (
	colorWhite   = lipgloss.Color("255")
	colorAccent  = lipgloss.Color("140")
	colorMuted   = lipgloss.Color("244") // серый — неактивные элементы
	colorError   = lipgloss.Color("203")
	colorSuccess = lipgloss.Color("150")
)

var (
	labelStyle = lipgloss.NewStyle().Foreground(colorMuted)
	errorStyle = lipgloss.NewStyle().Foreground(colorError)
	infoStyle  = lipgloss.NewStyle().Foreground(colorSuccess)

	inputBoxFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 1)

	inputBoxBlurred = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorAccent).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorMuted).
				Padding(0, 2)
)

func newStyledInput(placeholder string, isPass bool) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.SetWidth(25)

	if isPass {
		t.EchoMode = textinput.EchoPassword
		t.EchoCharacter = '•'
	}

	s := t.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(colorAccent)
	s.Focused.Text = lipgloss.NewStyle().Foreground(colorAccent)
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Prompt = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Text = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)
	s.Cursor.Color = colorAccent
	t.SetStyles(s)

	t.Prompt = "» "

	return t
}

func button(text, description string, disabled bool) string {
	textInfo := lipgloss.NewStyle().Faint(true).Render(description)
	if disabled {
		return lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Faint(true).
			Render(fmt.Sprintf("[%s]", text)) + " " + textInfo
	}
	return lipgloss.NewStyle().Foreground(colorAccent).Bold(true).
		Render(fmt.Sprintf("[%s]", text)) + " " + textInfo
}
