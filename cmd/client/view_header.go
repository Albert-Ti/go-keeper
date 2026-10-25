package main

import (
	"charm.land/lipgloss/v2"
)

func (m model) headerView() string {
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

	gapWidth := maxWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}
