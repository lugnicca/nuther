package views

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderDetails renders the details tab
func RenderDetails(drive smart.DriveInfo, width int, s *styles.Styles) string {
	var content strings.Builder

	// Drive details section
	content.WriteString(s.Bold.Foreground(s.AccentPrimary).Render(" 📋 Complete Details\n\n"))

	details := []struct{ label, value string }{
		{"Device Path", drive.Device},
		{"Model", drive.Model},
		{"Model Family", drive.ModelFamily},
		{"Serial Number", drive.Serial},
		{"Firmware Version", drive.Firmware},
		{"Capacity", drive.Capacity},
		{"Capacity (bytes)", smart.FormatNumber(drive.CapacityBytes)},
		{"Form Factor", drive.FormFactor},
		{"Interface", drive.Interface},
		{"Drive Type", drive.GetDriveType()},
		{"S.M.A.R.T. Support", smart.BoolToYesNo(drive.SmartSupported)},
		{"S.M.A.R.T. Enabled", smart.BoolToYesNo(drive.SmartEnabled)},
		{"Health Status", string(drive.HealthStatus)},
		{"Self-Test", smart.BoolToPassFail(drive.HealthPassed)},
		{"Temperature", s.FormatTemperature(drive.Temperature)},
		{"Power On Time", smart.FormatHours(drive.PowerOnHours)},
		{"Power Cycles", smart.FormatNumber(drive.PowerCycles)},
		{"Last Updated", drive.LastUpdate.Format("2006-01-02 15:04:05")},
	}

	if drive.RotationRate > 0 {
		details = append(details, struct{ label, value string }{
			"Rotation Rate", fmt.Sprintf("%d RPM", drive.RotationRate),
		})
	}

	for _, detail := range details {
		if detail.value != "" && detail.value != "0" {
			labelStyled := s.Dim.Render(components.PadRight(detail.label+":", 20))
			valueStyled := s.Bold.Render(detail.value)
			content.WriteString(fmt.Sprintf("  %s %s\n", labelStyled, valueStyled))
		}
	}

	// Critical attributes section for SATA drives
	if !drive.IsNVMe && len(drive.Attributes) > 0 {
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.Warning).Render(" "+styles.IconWarning+" Critical Attributes\n\n"))

		criticalIDs := []int{
			smart.AttrReallocatedSectors,
			smart.AttrReportedUncorrect,
			smart.AttrCommandTimeout,
			smart.AttrPendingSectors,
			smart.AttrOfflineUncorrectable,
			smart.AttrUDMACRCError,
		}

		for _, attr := range drive.Attributes {
			for _, id := range criticalIDs {
				if attr.ID == id {
					color := s.Success
					if attr.RawValue > 0 {
						color = s.Warning
					}
					if attr.RawValue > 100 {
						color = s.Danger
					}
					labelStyled := s.Dim.Render(components.PadRight(attr.Name+":", 28))
					valueStyled := lipgloss.NewStyle().Bold(true).Foreground(color).Render(attr.RawString)
					content.WriteString(fmt.Sprintf("  %s %s\n", labelStyled, valueStyled))
				}
			}
		}
	}

	return components.RenderBox(content.String(), width-4, "", s)
}
