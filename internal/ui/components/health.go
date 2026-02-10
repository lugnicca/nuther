package components

import (
	"nuther/internal/smart"

	"github.com/charmbracelet/lipgloss"
)

// RenderStyledStatus renders text with bold styling and the given foreground color
func RenderStyledStatus(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Bold(true).Foreground(color).Render(text)
}

// RenderHealthBadge renders a health status badge with icon, padded status text, and color
func RenderHealthBadge(status smart.HealthStatus, icon string, color lipgloss.Color, width int) string {
	return lipgloss.NewStyle().Bold(true).Foreground(color).Render(icon + " " + PadRight(string(status), width))
}
