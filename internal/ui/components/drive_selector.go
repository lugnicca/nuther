package components

import (
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

const maxDriveNameLength = 28

// RenderDriveSelector renders the drive selector buttons
func RenderDriveSelector(drives []smart.DriveInfo, selected int, s *styles.Styles) string {
	var buttons []string

	for i, drive := range drives {
		// Health indicator
		healthColor := s.GetHealthColor(drive.HealthStatus)
		healthIcon := lipgloss.NewStyle().Foreground(healthColor).Render(styles.IconBullet)

		// Drive name (truncated)
		name := Truncate(drive.Model, maxDriveNameLength)

		// Build button content
		buttonContent := healthIcon + " " + name

		// Style based on selection
		var button string
		if i == selected {
			button = s.DriveButtonSelected.Render(buttonContent)
		} else {
			button = s.DriveButton.Render(buttonContent)
		}

		buttons = append(buttons, button)
	}

	return strings.Join(buttons, " ")
}

