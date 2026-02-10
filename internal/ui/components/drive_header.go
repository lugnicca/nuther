package components

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderDriveHeader renders the drive header with model, capacity, firmware, serial, health
func RenderDriveHeader(drive smart.DriveInfo, s *styles.Styles) string {
	var header strings.Builder

	healthColor := s.GetHealthColor(drive.HealthStatus)
	healthIcon := s.GetHealthIcon(drive.HealthStatus)
	driveIcon := s.GetDriveIcon(&drive)

	// Calculate health percentage
	healthPercent := 100
	if drive.IsNVMe && drive.NVMeHealthLog != nil {
		healthPercent = 100 - drive.NVMeHealthLog.PercentageUsed
		if healthPercent < 0 {
			healthPercent = 0
		}
	}

	// First line: Icon + Model + Health badge
	header.WriteString(" ")
	header.WriteString(lipgloss.NewStyle().Foreground(s.AccentPrimary).Render(driveIcon))
	header.WriteString("  ")
	header.WriteString(s.Bold.Foreground(s.Text).Render(drive.Model))
	header.WriteString("   ")

	// Health badge
	healthBadge := fmt.Sprintf(" %s %s ", healthIcon, drive.HealthStatus)
	header.WriteString(s.GetHealthStyle(drive.HealthStatus).Render(healthBadge))
	header.WriteString("\n")

	// Second line: Capacity, Firmware, Serial
	header.WriteString("    ")
	header.WriteString(lipgloss.NewStyle().Foreground(s.AccentPrimary).Render(drive.Capacity))
	header.WriteString("   ")
	header.WriteString(s.Dim.Render("FW:"))
	header.WriteString(" ")
	header.WriteString(s.Base.Render(drive.Firmware))
	header.WriteString("   ")
	header.WriteString(s.Dim.Render("SN:"))
	header.WriteString(" ")
	header.WriteString(s.Base.Render(drive.Serial))
	header.WriteString("\n")

	// Health ring (text-based)
	header.WriteString("\n")
	borderStyle := lipgloss.NewStyle().Foreground(healthColor)
	healthText := fmt.Sprintf("  %s %3d%%  ", healthIcon, healthPercent)

	header.WriteString("    ")
	header.WriteString(borderStyle.Render("╭──────────╮"))
	header.WriteString("      ")

	// Protocol info labels
	header.WriteString(s.Dim.Render("PROTOCOL        "))
	header.WriteString(s.Dim.Render("ROTATION        "))
	header.WriteString(s.Dim.Render("FORM FACTOR     "))
	header.WriteString("\n")

	header.WriteString("    ")
	header.WriteString(borderStyle.Render("│"))
	header.WriteString(lipgloss.NewStyle().Bold(true).Foreground(healthColor).Render(healthText))
	header.WriteString(borderStyle.Render("│"))
	header.WriteString("      ")

	// Protocol value
	protocol := drive.Interface
	if drive.IsNVMe {
		protocol = "NVMe"
	} else if drive.IsSSD {
		protocol = "SATA SSD"
	}
	header.WriteString(s.Bold.Render(PadRight(protocol, 16)))

	// Rotation value
	rotation := "0 RPM (SSD)"
	if drive.RotationRate > 0 {
		rotation = fmt.Sprintf("%d RPM", drive.RotationRate)
	}
	header.WriteString(s.Bold.Render(PadRight(rotation, 16)))

	// Form factor value
	header.WriteString(s.Bold.Render(drive.FormFactor))
	header.WriteString("\n")

	header.WriteString("    ")
	header.WriteString(borderStyle.Render("╰──────────╯"))
	header.WriteString("\n")

	return header.String()
}
