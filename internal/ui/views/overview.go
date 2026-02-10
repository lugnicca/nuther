package views

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderOverview renders the overview tab
func RenderOverview(drive smart.DriveInfo, selectedAttr, scrollOffset, width, height int, s *styles.Styles) string {
	var result strings.Builder

	// Drive header with model, capacity, firmware, serial, health
	result.WriteString(components.RenderDriveHeader(drive, s))
	result.WriteString("\n")

	// Stats cards (temp, hours, cycles, interface)
	result.WriteString(components.RenderStatsCards(drive, s))
	result.WriteString("\n")

	// Attributes table
	result.WriteString(renderAttributesTable(drive, selectedAttr, scrollOffset, width, height, s))

	return result.String()
}

func renderAttributesTable(drive smart.DriveInfo, selectedAttr, scrollOffset, width, height int, s *styles.Styles) string {
	if drive.IsNVMe {
		return renderNVMeAttributesCompact(drive, selectedAttr, s)
	}

	if len(drive.Attributes) == 0 {
		return s.Dim.Render("  No S.M.A.R.T. attributes available\n")
	}

	var content strings.Builder

	// Header
	content.WriteString(" ")
	content.WriteString(s.Bold.Foreground(s.AccentPrimary).Render("≡ S.M.A.R.T. Attributes"))
	content.WriteString("\n\n")

	// Table header
	header := fmt.Sprintf(" %s  %-4s  %-28s  %-8s  %-8s  %-8s  %-8s  %s",
		" ", "ID", "NAME", "FAILED", "VALUE", "WORST", "THRESH", "RAW")
	content.WriteString(s.Dim.Render(header))
	content.WriteString("\n")
	content.WriteString(" " + components.RenderHorizontalLine(width-6, s))
	content.WriteString("\n")

	// Sort attributes by ID
	attrs := components.SortAttributesByID(drive.Attributes)

	// Calculate visible range
	// uiChrome: vertical space consumed by header + tabs + drive selector + status bar + margins
	const uiChrome = 22
	maxVisible := height - uiChrome
	if maxVisible < 5 {
		maxVisible = 5
	}
	if maxVisible > len(attrs) {
		maxVisible = len(attrs)
	}

	visibleStart := scrollOffset
	visibleEnd := visibleStart + maxVisible
	if visibleEnd > len(attrs) {
		visibleEnd = len(attrs)
		visibleStart = visibleEnd - maxVisible
		if visibleStart < 0 {
			visibleStart = 0
		}
	}

	// Render visible attributes
	for i := visibleStart; i < visibleEnd; i++ {
		attr := attrs[i]

		// Status icon
		status := attr.GetStatus()
		statusColor := s.GetHealthColor(status)
		statusIcon := lipgloss.NewStyle().Foreground(statusColor).Render(s.GetHealthIcon(status))

		// When failed
		whenFailed := "never"
		if attr.WhenFailed != "" {
			whenFailed = attr.WhenFailed
		}

		// Raw value
		rawStr := components.FormatRawValue(attr)

		rowStyle := components.GetTableRowStyle(i, selectedAttr, s)

		// Build row
		row := fmt.Sprintf(" %s  %-4d  %-28s  %-8s  %-8d  %-8d  %-8d  %s",
			statusIcon,
			attr.ID,
			components.Truncate(attr.Name, 28),
			whenFailed,
			attr.Value,
			attr.Worst,
			attr.Threshold,
			rawStr)

		content.WriteString(rowStyle.Render(row))
		content.WriteString("\n")
	}

	// Scroll indicator
	if len(attrs) > maxVisible {
		scrollInfo := fmt.Sprintf("\n  ↑↓ Scroll [%d-%d / %d]", visibleStart+1, visibleEnd, len(attrs))
		content.WriteString(s.Dim.Render(scrollInfo))
	}

	return content.String()
}

func renderNVMeAttributesCompact(drive smart.DriveInfo, selectedAttr int, s *styles.Styles) string {
	if len(drive.NVMeAttributes) == 0 {
		return s.Dim.Render("  No NVMe attributes available\n")
	}

	var content strings.Builder

	// Header
	content.WriteString(" ")
	content.WriteString(s.Bold.Foreground(s.AccentPrimary).Render("≡ NVMe Attributes"))
	content.WriteString("\n\n")

	// Table header
	header := fmt.Sprintf(" %s  %-28s  %-20s  %s",
		" ", "ATTRIBUTE", "VALUE", "STATUS")
	content.WriteString(s.Dim.Render(header))
	content.WriteString("\n")
	content.WriteString(" " + components.RenderHorizontalLine(70, s))
	content.WriteString("\n")

	// Render attributes
	for i, attr := range drive.NVMeAttributes {
		// Status icon
		statusColor := s.GetHealthColor(attr.Status)
		statusIcon := lipgloss.NewStyle().Foreground(statusColor).Render(s.GetHealthIcon(attr.Status))

		// Row styling
		rowStyle := components.GetTableRowStyle(i, selectedAttr, s)

		statusStyled := components.RenderStyledStatus(string(attr.Status), statusColor)
		row := fmt.Sprintf(" %s  %-28s  %-20s  %s",
			statusIcon,
			attr.Name,
			attr.RawValue,
			statusStyled)

		content.WriteString(rowStyle.Render(row))
		content.WriteString("\n")
	}

	return content.String()
}
