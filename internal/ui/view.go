package ui

import (
	"strings"

	"nuther/internal/ui/components"
	"nuther/internal/ui/views"
)

// View implements tea.Model
func (m Model) View() string {
	if !m.Ready {
		return m.Spinner.View() + " Initializing..."
	}

	var sections []string

	// Header with logo
	sections = append(sections, components.RenderHeader(m.Styles, m.Config))

	// Check if we have drives
	if len(m.Drives) == 0 {
		warningBox := components.RenderBox(
			"No drives detected.\n\n"+
				"Please run with sudo to access S.M.A.R.T. data:\n"+
				"  sudo nuther\n\n"+
				"Demo mode is active with sample data.",
			m.Width-4, "", m.Styles)
		sections = append(sections, warningBox)
		return strings.Join(sections, "\n")
	}

	// Check if in demo mode
	isDemo := len(m.Drives) > 0 && m.Drives[0].IsDemo

	// Demo mode banner
	if isDemo {
		sections = append(sections, components.RenderDemoBanner(m.Width, m.Styles))
	}

	// Drive selector (not shown in settings)
	if m.ActiveTab != TabSettings {
		sections = append(sections, components.RenderDriveSelector(m.Drives, m.SelectedDrive, m.Styles))
		sections = append(sections, "")
	}

	// Tab bar
	sections = append(sections, components.RenderTabs(m.Tabs, m.ActiveTab, m.Width, m.Styles))

	// Main content based on active tab
	switch m.ActiveTab {
	case TabAllDrives:
		sections = append(sections, views.RenderAllDrives(m.Drives, m.SelectedDrive, m.Width, m.Height, m.Styles))
	case TabSectorGrid:
		sections = append(sections, views.RenderSectorGrid(m.Drives, m.SelectedDrive, m.Width, m.Height, m.Styles))
	case TabSettings:
		sections = append(sections, views.RenderSettings(m.Config, m.SettingsSelected, m.SettingsMessage, m.Width, m.Styles))
	default:
		drive := m.GetCurrentDrive()
		if drive != nil {
			switch m.ActiveTab {
			case TabOverview:
				sections = append(sections, views.RenderOverview(*drive, m.SelectedAttr, m.ScrollOffset, m.Width, m.Height, m.Styles))
			case TabAttributes:
				sections = append(sections, views.RenderAttributes(*drive, m.SelectedAttr, m.ScrollOffset, m.Width, m.Height, m.Styles))
			case TabDetails:
				sections = append(sections, views.RenderDetails(*drive, m.Width, m.Styles))
			}
		}
	}

	// Screenshot status notification
	if m.ScreenshotStatus != "" {
		sections = append(sections, components.RenderScreenshotStatus(m.ScreenshotStatus, m.ScreenshotMessage, m.Styles))
	}

	// Status bar
	sections = append(sections, "")
	sections = append(sections, components.RenderStatusBar(m.SelectedDrive, len(m.Drives), m.LastRefresh, isDemo, m.isCacheFresh(), m.Width, m.KeyMap, m.Styles))

	// Help overlay if shown
	if m.ShowHelp {
		helpContent := components.RenderHelp(m.Styles)
		return m.overlayHelp(strings.Join(sections, "\n"), helpContent)
	}

	return strings.Join(sections, "\n")
}

// overlayHelp places the help overlay on top of the screen
func (m Model) overlayHelp(screen, helpContent string) string {
	lines := strings.Split(screen, "\n")
	helpLines := strings.Split(helpContent, "\n")

	// Position help in the center
	startY := (m.Height - len(helpLines)) / 2
	if startY < 0 {
		startY = 0
	}

	for i, helpLine := range helpLines {
		lineIdx := startY + i
		if lineIdx < len(lines) {
			// Simple overlay - just replace the line
			lines[lineIdx] = helpLine
		}
	}

	return strings.Join(lines, "\n")
}
