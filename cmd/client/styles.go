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
	colorPrimary     = lipgloss.Color("212") // розовый — акцент на активном поле
	colorMuted       = lipgloss.Color("240") // серый — неактивные элементы
	colorPlaceholder = lipgloss.Color("244")
	colorError       = lipgloss.Color("203")
	colorText        = lipgloss.Color("255")
)

var (
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	errorStyle = lipgloss.NewStyle().Foreground(colorError)

	inputBoxFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	inputBoxBlurred = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorPrimary).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle().
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
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(colorPrimary)
	s.Focused.Text = lipgloss.NewStyle().Foreground(colorPrimary)
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(colorPlaceholder)
	s.Blurred.Prompt = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Text = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)
	s.Cursor.Color = colorPrimary
	t.SetStyles(s)

	t.Prompt = "» "

	return t
}

func button(text, description string, disabled bool) string {
	textInfo := lipgloss.NewStyle().Faint(true).Render(description)
	if disabled {
		return lipgloss.NewStyle().Foreground(colorPrimary).Faint(true).Render(fmt.Sprintf("[ %s ]", text)) + " " + textInfo
	}
	return lipgloss.NewStyle().Foreground(colorPrimary).Render(fmt.Sprintf("[ %s ]", text)) + " " + textInfo
}
