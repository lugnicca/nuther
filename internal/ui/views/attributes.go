package views

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

const maxVisibleAttributes = 14

// RenderAttributes renders the attributes tab
func RenderAttributes(drive smart.DriveInfo, selectedAttr, scrollOffset, width, height int, s *styles.Styles) string {
	if drive.IsNVMe {
		return renderNVMeAttributes(drive, selectedAttr, width, s)
	}
	return renderSATAAttributes(drive, selectedAttr, scrollOffset, width, s)
}

func renderSATAAttributes(drive smart.DriveInfo, selectedAttr, scrollOffset, width int, s *styles.Styles) string {
	if len(drive.Attributes) == 0 {
		return components.RenderBox("No S.M.A.R.T. attributes available", width-4, "", s)
	}

	var content strings.Builder

	// Header
	header := fmt.Sprintf("%-4s %-22s %6s %6s %6s %10s %6s",
		"ID", "Attribute", "Value", "Worst", "Thresh", "Raw", "Status")
	content.WriteString(s.TableHeader.Render(header))
	content.WriteString("\n")
	content.WriteString(components.RenderHorizontalLine(70, s))
	content.WriteString("\n")

	// Sort attributes by ID
	attrs := components.SortAttributesByID(drive.Attributes)

	// Calculate visible range
	visibleStart := scrollOffset
	visibleEnd := visibleStart + maxVisibleAttributes
	if visibleEnd > len(attrs) {
		visibleEnd = len(attrs)
	}

	// Render visible attributes
	for i := visibleStart; i < visibleEnd; i++ {
		attr := attrs[i]

		// Determine status
		status := attr.GetStatus()
		statusColor := s.GetHealthColor(status)

		// Format raw string
		rawStr := components.FormatRawValue(attr)
		if len(rawStr) > 10 {
			rawStr = rawStr[:10]
		}

		rowStyle := components.GetTableRowStyle(i, selectedAttr, s)

		// Build row
		statusStyled := components.RenderStyledStatus(components.PadRight(string(status), 6), statusColor)
		row := fmt.Sprintf("%-4d %-22s %6d %6d %6d %10s %s",
			attr.ID,
			components.Truncate(attr.Name, 22),
			attr.Value,
			attr.Worst,
			attr.Threshold,
			rawStr,
			statusStyled)

		content.WriteString(rowStyle.Render(row))
		content.WriteString("\n")
	}

	// Description bar for selected attribute
	if selectedAttr >= 0 && selectedAttr < len(attrs) {
		if desc, ok := smart.AttributeDescriptions[attrs[selectedAttr].ID]; ok {
			content.WriteString("\n")
			content.WriteString(s.Italic.Foreground(s.TextDim).Render(fmt.Sprintf("%s %s", styles.IconInfo, desc)))
		}
	}

	// Legend
	content.WriteString("\n")
	content.WriteString(s.Italic.Foreground(s.TextDim).Render("Value: Current (100=optimal) • Thresh: Failure threshold • Raw: Actual count"))

	// Scroll indicator
	if len(attrs) > maxVisibleAttributes {
		scrollInfo := fmt.Sprintf(" [%d-%d of %d]", visibleStart+1, visibleEnd, len(attrs))
		content.WriteString(lipgloss.NewStyle().Foreground(s.AccentPrimary).Render(scrollInfo))
	}

	return components.RenderBox(content.String(), width-4, "S.M.A.R.T. Attributes", s)
}

func renderNVMeAttributes(drive smart.DriveInfo, selectedAttr, width int, s *styles.Styles) string {
	if len(drive.NVMeAttributes) == 0 {
		return components.RenderBox("No NVMe attributes available", width-4, "", s)
	}

	var content strings.Builder

	// Header
	header := fmt.Sprintf("%-26s %-18s %s", "Attribute", "Value", "Status")
	content.WriteString(s.TableHeader.Render(header))
	content.WriteString("\n")
	content.WriteString(components.RenderHorizontalLine(52, s))
	content.WriteString("\n")

	// Render attributes
	for i, attr := range drive.NVMeAttributes {
		statusColor := s.GetHealthColor(attr.Status)

		rowStyle := components.GetTableRowStyle(i, selectedAttr, s)

		statusStyled := components.RenderStyledStatus(string(attr.Status), statusColor)
		row := fmt.Sprintf("%-26s %-18s %s",
			attr.Name,
			attr.RawValue,
			statusStyled)

		content.WriteString(rowStyle.Render(row))
		content.WriteString("\n")
	}

	// Description bar for selected attribute
	if selectedAttr >= 0 && selectedAttr < len(drive.NVMeAttributes) {
		desc := drive.NVMeAttributes[selectedAttr].Description
		if desc != "" {
			content.WriteString("\n")
			content.WriteString(s.Italic.Foreground(s.TextDim).Render(fmt.Sprintf("%s %s", styles.IconInfo, desc)))
		}
	}

	return components.RenderBox(content.String(), width-4, "NVMe Health Attributes", s)
}
