package views

import (
	"fmt"
	"strings"

	"nuther/internal/config"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)


// RenderSettings renders the settings tab
func RenderSettings(cfg *config.Config, selected int, message string, width int, s *styles.Styles) string {
	var result strings.Builder

	result.WriteString("\n")
	result.WriteString(" ")
	result.WriteString(s.Bold.Foreground(s.AccentPrimary).Render("Settings"))
	result.WriteString("\n")
	result.WriteString(" ")
	result.WriteString(s.Dim.Render("Use ↑/↓ to select, ←/→ to change, Enter to save"))
	result.WriteString("\n\n")

	borderStyle := lipgloss.NewStyle().Foreground(s.Border)

	// Theme setting
	result.WriteString(renderSettingRow(
		"Theme",
		cfg.Theme,
		getThemeOptions(),
		selected == 0,
		s,
		borderStyle,
	))
	result.WriteString("\n")

	// Show Logo setting
	showLogoValue := "Yes"
	if !cfg.Display.ShowLogo {
		showLogoValue = "No"
	}
	result.WriteString(renderSettingRow(
		"Show Logo",
		showLogoValue,
		[]string{"Yes", "No"},
		selected == 1,
		s,
		borderStyle,
	))
	result.WriteString("\n")

	// Temperature Unit setting
	tempUnit := "Celsius"
	if cfg.Display.ShowFahrenheit {
		tempUnit = "Fahrenheit"
	}
	result.WriteString(renderSettingRow(
		"Temperature",
		tempUnit,
		[]string{"Celsius", "Fahrenheit"},
		selected == 2,
		s,
		borderStyle,
	))

	// Status message
	if message != "" {
		result.WriteString("\n\n ")
		successStyle := lipgloss.NewStyle().Bold(true).Foreground(s.Success)
		result.WriteString(successStyle.Render("✓ " + message))
	}

	// Help text
	result.WriteString("\n\n ")
	result.WriteString(s.Dim.Render("Press Enter to save settings to config file"))
	result.WriteString("\n ")
	configPath, _ := config.GetConfigPath()
	result.WriteString(s.Dim.Render(fmt.Sprintf("Config: %s", configPath)))

	return result.String()
}

func renderSettingRow(label, currentValue string, options []string, isSelected bool, s *styles.Styles, borderStyle lipgloss.Style) string {
	var row strings.Builder

	// Selection indicator
	if isSelected {
		row.WriteString(" ")
		row.WriteString(lipgloss.NewStyle().Foreground(s.AccentPrimary).Bold(true).Render("▶"))
		row.WriteString(" ")
	} else {
		row.WriteString("   ")
	}

	// Label
	labelStyle := s.Base
	if isSelected {
		labelStyle = s.Bold.Foreground(s.AccentPrimary)
	}
	row.WriteString(labelStyle.Render(components.PadRight(label, 16)))

	// Value with arrows if selected
	if isSelected {
		row.WriteString(" ")
		row.WriteString(lipgloss.NewStyle().Foreground(s.AccentSecondary).Render("◀"))
		row.WriteString(" ")
		row.WriteString(s.Bold.Foreground(lipgloss.Color("#ffffff")).Render(components.PadCenter(currentValue, 14)))
		row.WriteString(" ")
		row.WriteString(lipgloss.NewStyle().Foreground(s.AccentSecondary).Render("▶"))
	} else {
		row.WriteString("   ")
		row.WriteString(s.Dim.Render(components.PadCenter(currentValue, 14)))
		row.WriteString("  ")
	}

	return row.String()
}

func getThemeOptions() []string {
	return config.ListThemes()
}
