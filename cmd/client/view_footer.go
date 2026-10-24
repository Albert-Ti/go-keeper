package main

import (
	"charm.land/lipgloss/v2"
)

func footerView() string {
	leftRendered := lipgloss.NewStyle().Faint(true).
		Render("build " + VERSION)

	rightRendered := lipgloss.NewStyle().Foreground(colorAccent).Render("© Albert Taygibov")

	gapWidth := maxWidth - lipgloss.Width(leftRendered) - lipgloss.Width(rightRendered)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, gap, rightRendered)
}
