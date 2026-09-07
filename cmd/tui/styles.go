package main

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

const cardWidth = 80

// Color
var (
	colorPrimary     = lipgloss.Color("212") // розовый — акцент на активном поле
	colorMuted       = lipgloss.Color("240") // серый — неактивные элементы
	colorPlaceholder = lipgloss.Color("244")
	colorError       = lipgloss.Color("203")
	colorText        = lipgloss.Color("255")
)

// Component UI Style
var (
	menuStyle  = lipgloss.NewStyle().Foreground(colorText)
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
)

func newStyledInput(placeholder string, isPassword bool) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.SetWidth(30)

	if isPassword {
		t.EchoMode = textinput.EchoPassword
		t.EchoCharacter = '•'
	}

	s := t.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(colorPrimary)
	s.Focused.Text = lipgloss.NewStyle().Foreground(colorText)
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(colorPlaceholder)
	s.Blurred.Prompt = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Text = lipgloss.NewStyle().Foreground(colorMuted)
	s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)
	s.Cursor.Color = colorPrimary
	t.SetStyles(s)

	t.Prompt = "» "

	return t
}

func divider(width int) string {
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
