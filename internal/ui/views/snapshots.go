package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"nuther/internal/smartwatch"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

const snapshotTimeFormat = "2006-01-02 15:04:05"

// RenderSnapshots renders the archived SMART snapshot history tab.
func RenderSnapshots(index smartwatch.Index, selectedSnapshot, width, height int, s *styles.Styles) string {
	contentWidth := width - 8
	if contentWidth < 60 {
		contentWidth = 60
	}

	if len(index.Devices) == 0 && len(index.Snapshots) == 0 {
		content := strings.Join([]string{
			"No SMART snapshots archived yet.",
			"",
			"Run `sudo nuther watch-smart --once` to capture current drives,",
			"or `sudo nuther watch-smart` to keep watching for newly connected drives.",
		}, "\n")
		return components.RenderBox(content, width-4, "Snapshots", s)
	}

	records := sortedSnapshotRecords(index.Snapshots)
	if selectedSnapshot < 0 || selectedSnapshot >= len(records) {
		selectedSnapshot = 0
	}

	var content strings.Builder
	content.WriteString(renderSnapshotSummary(index, records, s))
	content.WriteString("\n\n")
	content.WriteString(renderSnapshotDevices(index, contentWidth, s))
	content.WriteString("\n\n")
	content.WriteString(renderSnapshotHistory(records, selectedSnapshot, contentWidth, height, s))

	return components.RenderBox(content.String(), width-4, "Snapshots", s)
}

func sortedSnapshotRecords(records []smartwatch.SnapshotRecord) []smartwatch.SnapshotRecord {
	sorted := append([]smartwatch.SnapshotRecord(nil), records...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})
	return sorted
}

func sortedDeviceRecords(devices map[string]smartwatch.DeviceRecord) []smartwatch.DeviceRecord {
	sorted := make([]smartwatch.DeviceRecord, 0, len(devices))
	for _, device := range devices {
		sorted = append(sorted, device)
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		if !sorted[i].LastSeen.Equal(sorted[j].LastSeen) {
			return sorted[i].LastSeen.After(sorted[j].LastSeen)
		}
		return snapshotDeviceLabel(sorted[i].Summary) < snapshotDeviceLabel(sorted[j].Summary)
	})
	return sorted
}

func renderSnapshotSummary(index smartwatch.Index, records []smartwatch.SnapshotRecord, s *styles.Styles) string {
	updated := "never"
	if !index.UpdatedAt.IsZero() {
		updated = formatSnapshotTime(index.UpdatedAt)
	} else if len(records) > 0 {
		updated = formatSnapshotTime(records[0].Timestamp)
	}

	return fmt.Sprintf("  %s known disks   %s snapshots   updated %s",
		s.Bold.Render(fmt.Sprintf("%d", len(index.Devices))),
		s.Bold.Render(fmt.Sprintf("%d", len(records))),
		s.Dim.Render(updated),
	)
}

func renderSnapshotDevices(index smartwatch.Index, contentWidth int, s *styles.Styles) string {
	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.AccentSecondary).Render(" Known disks"))
	content.WriteString("\n")
	content.WriteString(components.RenderHorizontalLine(contentWidth, s))
	content.WriteString("\n")

	devices := sortedDeviceRecords(index.Devices)
	if len(devices) == 0 {
		content.WriteString(s.Dim.Render("  No device records yet."))
		return content.String()
	}

	header := fmt.Sprintf("%-3s %-26s %-9s %-19s %-19s %-10s",
		"#", "Disk", "Health", "First seen", "Last seen", "Snapshots")
	content.WriteString(s.TableHeader.Render(header))
	content.WriteString("\n")

	limit := len(devices)
	if limit > 6 {
		limit = 6
	}
	for i := 0; i < limit; i++ {
		device := devices[i]
		healthColor := s.GetHealthColor(device.Summary.HealthStatus)
		healthIcon := s.GetHealthIcon(device.Summary.HealthStatus)
		health := components.RenderHealthBadge(device.Summary.HealthStatus, healthIcon, healthColor, 7)
		row := fmt.Sprintf("%-3d %-26s %s %-19s %-19s %-10d",
			i+1,
			components.Truncate(snapshotDeviceLabel(device.Summary), 26),
			health,
			formatSnapshotTime(device.FirstSeen),
			formatSnapshotTime(device.LastSeen),
			len(device.SnapshotIDs),
		)
		content.WriteString(components.GetTableRowStyle(i, -1, s).Render(row))
		content.WriteString("\n")
	}
	if len(devices) > limit {
		content.WriteString(s.Dim.Render(fmt.Sprintf("  +%d older disk records", len(devices)-limit)))
	}
	return strings.TrimRight(content.String(), "\n")
}

func renderSnapshotHistory(records []smartwatch.SnapshotRecord, selectedSnapshot, contentWidth, height int, s *styles.Styles) string {
	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.AccentSecondary).Render(" Snapshot history"))
	content.WriteString("\n")
	content.WriteString(components.RenderHorizontalLine(contentWidth, s))
	content.WriteString("\n")

	if len(records) == 0 {
		content.WriteString(s.Dim.Render("  No snapshot records yet."))
		return content.String()
	}

	header := fmt.Sprintf("%-3s %-19s %-12s %-26s %-9s %-28s",
		"#", "Captured", "Reason", "Disk", "Health", "Snapshot ID")
	content.WriteString(s.TableHeader.Render(header))
	content.WriteString("\n")

	visible := height - 28
	if visible < 4 {
		visible = 4
	}
	if visible > 10 {
		visible = 10
	}
	if visible > len(records) {
		visible = len(records)
	}
	start := selectedSnapshot - visible/2
	if start < 0 {
		start = 0
	}
	if start+visible > len(records) {
		start = len(records) - visible
	}
	if start < 0 {
		start = 0
	}

	for i := start; i < start+visible; i++ {
		record := records[i]
		healthColor := s.GetHealthColor(record.Device.HealthStatus)
		healthIcon := s.GetHealthIcon(record.Device.HealthStatus)
		health := components.RenderHealthBadge(record.Device.HealthStatus, healthIcon, healthColor, 7)
		row := fmt.Sprintf("%-3d %-19s %-12s %-26s %s %-28s",
			i+1,
			formatSnapshotTime(record.Timestamp),
			components.Truncate(record.Reason, 12),
			components.Truncate(snapshotDeviceLabel(record.Device), 26),
			health,
			components.Truncate(record.ID, 28),
		)
		content.WriteString(components.GetTableRowStyle(i, selectedSnapshot, s).Render(row))
		content.WriteString("\n")
	}

	if len(records) > visible {
		content.WriteString(s.Dim.Render(fmt.Sprintf("  Showing %d-%d of %d snapshots · j/k scroll", start+1, start+visible, len(records))))
		content.WriteString("\n")
	}
	selected := records[selectedSnapshot]
	content.WriteString(s.Dim.Render(fmt.Sprintf("  Selected: Enter: open overview · GET /snapshots/%s · file %s", selected.ID, selected.Path)))

	return strings.TrimRight(content.String(), "\n")
}

func formatSnapshotTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format(snapshotTimeFormat)
}

func snapshotDeviceLabel(device smartwatch.DeviceSummary) string {
	for _, value := range []string{device.Model, device.Serial, device.Device, device.Key} {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "Unknown disk"
}
