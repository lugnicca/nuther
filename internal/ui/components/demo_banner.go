package components

import (
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderDemoBanner renders a warning banner indicating demo mode is active
func RenderDemoBanner(width int, s *styles.Styles) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Warning).
		Width(width).
		Align(lipgloss.Center)

	return style.Render(styles.IconWarning + " MODE DEMO — Donnees fictives. Lancez avec sudo pour les vraies donnees S.M.A.R.T.")
}
