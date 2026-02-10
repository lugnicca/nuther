package views

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderAllDrives renders the all drives tab
func RenderAllDrives(drives []smart.DriveInfo, selectedDrive, width, height int, s *styles.Styles) string {
	if len(drives) == 0 {
		return components.RenderBox("No drives detected.", width-4, "", s)
	}

	var content strings.Builder

	// Header row
	header := fmt.Sprintf("%-3s %-24s %-10s %-8s %-6s %-14s %-10s",
		"#", "Model", "Capacity", "Health", "Temp", "Power On", "Interface")
	content.WriteString(s.TableHeader.Render(header))
	content.WriteString("\n")
	content.WriteString(components.RenderHorizontalLine(width-8, s))
	content.WriteString("\n")

	// Drive rows
	for i, drive := range drives {
		healthColor := s.GetHealthColor(drive.HealthStatus)
		healthIcon := s.GetHealthIcon(drive.HealthStatus)
		tempColor := s.GetTemperatureColor(drive.Temperature)

		rowStyle := components.GetTableRowStyle(i, selectedDrive, s)

		// Format each column
		num := fmt.Sprintf("%d", i+1)
		model := components.Truncate(drive.Model, 24)
		capacity := components.PadRight(drive.Capacity, 10)
		healthStyled := components.RenderHealthBadge(drive.HealthStatus, healthIcon, healthColor, 6)
		tempStyled := lipgloss.NewStyle().Foreground(tempColor).Render(components.PadRight(s.FormatTemperature(drive.Temperature), 6))
		powerOn := components.PadRight(smart.FormatHours(drive.PowerOnHours), 14)
		iface := components.Truncate(drive.Interface, 10)

		row := fmt.Sprintf("%-3s %-24s %-10s %s %s %-14s %-10s",
			num, model, capacity, healthStyled, tempStyled, powerOn, iface)

		content.WriteString(rowStyle.Render(row))
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Summary section
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.AccentSecondary).Render(" Summary\n"))
	content.WriteString(components.RenderHorizontalLine(40, s))
	content.WriteString("\n")

	// Count drives by health status
	goodCount, cautionCount, badCount := 0, 0, 0
	var totalCapacity int64
	for _, drive := range drives {
		switch drive.HealthStatus {
		case smart.HealthGood:
			goodCount++
		case smart.HealthCaution:
			cautionCount++
		case smart.HealthBad:
			badCount++
		}
		totalCapacity += drive.CapacityBytes
	}

	totalDrivesStyled := s.Bold.Render(fmt.Sprintf("%d", len(drives)))
	totalCapacityStyled := s.Bold.Render(smart.FormatBytes(totalCapacity))

	content.WriteString(fmt.Sprintf("  Total Drives: %s\n", totalDrivesStyled))
	content.WriteString(fmt.Sprintf("  Total Capacity: %s\n", totalCapacityStyled))

	goodStyled := lipgloss.NewStyle().Foreground(s.Success).Render(styles.IconBullet)
	cautionStyled := lipgloss.NewStyle().Foreground(s.Warning).Render(styles.IconBullet)
	badStyled := lipgloss.NewStyle().Foreground(s.Danger).Render(styles.IconBullet)

	content.WriteString(fmt.Sprintf("  %s Healthy: %d   %s Caution: %d   %s Bad: %d\n",
		goodStyled, goodCount,
		cautionStyled, cautionCount,
		badStyled, badCount))

	return components.RenderBox(content.String(), width-4, "All Drives Overview", s)
}

