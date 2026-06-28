package views

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/components"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

const (
	sectorGridMinWidth        = 36
	sectorGridPendingBadLimit = 10
	sectorGridMaxSide         = 24
	sectorGridMinSide         = 6
)

type sectorCellStatus int

const (
	sectorCellGood sectorCellStatus = iota
	sectorCellReallocated
	sectorCellPending
	sectorCellUncorrectable
)

// RenderSectorGrid renders a square bucket grid for the selected drive.
// Each cell is a virtual bucket of the selected disk surface. S.M.A.R.T. does not
// expose a physical LBA map, so sector counters are distributed deterministically
// across buckets to make the amount and severity of damage immediately visible.
func RenderSectorGrid(drives []smart.DriveInfo, selectedDrive, width, height int, s *styles.Styles) string {
	if len(drives) == 0 {
		return components.RenderBox("No drives detected.", width-4, "", s)
	}

	if selectedDrive < 0 || selectedDrive >= len(drives) {
		selectedDrive = 0
	}

	contentWidth := width - 8
	if contentWidth < sectorGridMinWidth {
		contentWidth = sectorGridMinWidth
	}

	drive := drives[selectedDrive]
	status := sectorGridStatus(drive)
	wide := contentWidth >= 96
	side := sectorGridSide(contentWidth, height, wide)
	cells := sectorCellsForDrive(drive, side)

	var content strings.Builder
	content.WriteString(renderSelectedSectorHeader(drive, selectedDrive, len(drives), status, s))
	content.WriteString("\n")
	content.WriteString(s.Dim.Render("  virtual surface map · deterministic buckets, not physical LBA coordinates"))
	content.WriteString("\n\n")

	mapPanel := renderSectorMapPanel(cells, side, contentWidth, s)
	signalPanel := renderSectorSignalPanel(drive, status, side, contentWidth, s)
	if wide {
		content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, "  ", signalPanel))
	} else {
		content.WriteString(signalPanel)
		content.WriteString("\n\n")
		content.WriteString(mapPanel)
	}
	content.WriteString("\n")
	content.WriteString(s.Dim.Render(fmt.Sprintf("  %dx%d sector buckets · n/p switch disk", side, side)))

	return components.RenderBox(content.String(), width-4, "Sector Grid", s)
}

func sectorGridSide(contentWidth, terminalHeight int, wide bool) int {
	// Cells are rendered as two runes wide so they look closer to square in a terminal.
	reservedWidth := 4
	if wide {
		reservedWidth = 42
	}
	maxByWidth := (contentWidth - reservedWidth) / 2
	maxByHeight := terminalHeight - 20
	if wide {
		maxByHeight = terminalHeight - 18
	}
	if maxByHeight <= 0 {
		maxByHeight = sectorGridMinSide
	}
	side := minInt(maxByWidth, maxByHeight)
	if side > sectorGridMaxSide {
		side = sectorGridMaxSide
	}
	if side < sectorGridMinSide {
		side = sectorGridMinSide
	}
	return side
}

func renderSelectedSectorHeader(drive smart.DriveInfo, selectedDrive, totalDrives int, status smart.HealthStatus, s *styles.Styles) string {
	statusBadge := components.RenderHealthBadge(status, s.GetHealthIcon(status), s.GetHealthColor(status), 7)
	model := sectorGridDriveLabel(drive)
	return fmt.Sprintf("  Disk %d/%d  %s  %s", selectedDrive+1, totalDrives, statusBadge, s.Bold.Render(model))
}

func renderSectorMapPanel(cells []sectorCellStatus, side, contentWidth int, s *styles.Styles) string {
	grid := renderSectorCellGrid(cells, side, s)
	legend := renderSectorGridLegend(contentWidth, s)
	body := grid + "\n\n" + legend
	panelWidth := side*2 + 4
	if panelWidth < 34 {
		panelWidth = 34
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Border).
		Padding(1, 2).
		Width(panelWidth).
		Render(s.Bold.Foreground(s.AccentPrimary).Render("Surface map") + "\n" + body)
}

func renderSectorSignalPanel(drive smart.DriveInfo, status smart.HealthStatus, side, contentWidth int, s *styles.Styles) string {
	risk := sectorRiskScore(drive)
	riskColor := s.GetHealthColor(status)
	statusBadge := components.RenderHealthBadge(status, s.GetHealthIcon(status), s.GetHealthColor(status), 7)
	gauge := components.RenderGauge(risk, 100, 22, riskColor, s)

	var b strings.Builder
	b.WriteString(s.Bold.Foreground(s.AccentPrimary).Render("Disk signal"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s  Risk %s/100\n", statusBadge, lipgloss.NewStyle().Foreground(riskColor).Bold(true).Render(fmt.Sprintf("%d", risk))))
	b.WriteString(gauge)
	b.WriteString("\n\n")
	b.WriteString(renderMetricLine("Reallocated", drive.ReallocatedSectors, s.Warning, s))
	b.WriteString(renderMetricLine("Pending", drive.PendingSectors, s.AccentSecondary, s))
	b.WriteString(renderMetricLine("Uncorrectable", drive.UncorrectableSectors, s.Danger, s))
	b.WriteString(renderMetricLine("CRC/link", drive.CRCErrors, s.Info, s))
	b.WriteString("\n")
	b.WriteString(s.Dim.Render(fmt.Sprintf("Capacity %s", fallbackValue(drive.Capacity, "unknown"))))
	b.WriteString("\n")
	b.WriteString(s.Dim.Render(fmt.Sprintf("Temp %s · Power %s", s.FormatTemperature(drive.Temperature), smart.FormatHours(drive.PowerOnHours))))
	b.WriteString("\n\n")
	b.WriteString(renderSectorInterpretation(drive, status, s))

	panelWidth := 34
	if contentWidth < 72 {
		panelWidth = contentWidth - 4
	}
	if panelWidth < 30 {
		panelWidth = 30
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Border).
		Padding(1, 2).
		Width(panelWidth).
		Render(b.String())
}

func renderMetricLine(label string, value int64, color lipgloss.Color, s *styles.Styles) string {
	valueStyle := lipgloss.NewStyle().Foreground(color).Bold(value > 0)
	return fmt.Sprintf("%-14s %s\n", label, valueStyle.Render(smart.FormatNumber(value)))
}

func renderSectorInterpretation(drive smart.DriveInfo, status smart.HealthStatus, s *styles.Styles) string {
	switch status {
	case smart.HealthBad:
		return lipgloss.NewStyle().Foreground(s.Danger).Render("Replace / backup now")
	case smart.HealthCaution:
		return lipgloss.NewStyle().Foreground(s.Warning).Render("Watch this disk closely")
	case smart.HealthInfo:
		if drive.CRCErrors > 0 {
			return lipgloss.NewStyle().Foreground(s.Info).Render("Check cable / link stability")
		}
		return lipgloss.NewStyle().Foreground(s.Info).Render("Informational signal")
	default:
		return lipgloss.NewStyle().Foreground(s.Success).Render("No sector pressure")
	}
}

func renderSectorGridLegend(contentWidth int, s *styles.Styles) string {
	clean := lipgloss.NewStyle().Foreground(s.Success).Render("░░ clean")
	reallocated := lipgloss.NewStyle().Foreground(s.Warning).Render("██ reallocated")
	pending := lipgloss.NewStyle().Foreground(s.AccentSecondary).Render("██ pending")
	bad := lipgloss.NewStyle().Foreground(s.Danger).Render("██ uncorrectable")
	crc := lipgloss.NewStyle().Foreground(s.Info).Render("CRC/link = metric only")

	if contentWidth < 72 {
		return fmt.Sprintf("%s  %s\n%s  %s\n%s", clean, reallocated, pending, bad, crc)
	}
	return fmt.Sprintf("%s   %s\n%s   %s\n%s", clean, reallocated, pending, bad, crc)
}

func renderSectorCellGrid(cells []sectorCellStatus, side int, s *styles.Styles) string {
	var b strings.Builder

	for row := 0; row < side; row++ {
		for col := 0; col < side; col++ {
			idx := row*side + col
			b.WriteString(renderSectorCell(cells[idx], row, col, s))
		}
		if row < side-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func renderSectorCell(status sectorCellStatus, row, col int, s *styles.Styles) string {
	speckle := (row+col)%3 == 0
	switch status {
	case sectorCellUncorrectable:
		return lipgloss.NewStyle().Foreground(s.Danger).Render("▓▓")
	case sectorCellPending:
		return lipgloss.NewStyle().Foreground(s.AccentSecondary).Render("▒▒")
	case sectorCellReallocated:
		return lipgloss.NewStyle().Foreground(s.Warning).Render("▓▓")
	default:
		cell := "░░"
		if speckle {
			cell = "··"
		}
		return lipgloss.NewStyle().Foreground(s.Success).Render(cell)
	}
}

func sectorCellsForDrive(drive smart.DriveInfo, side int) []sectorCellStatus {
	total := side * side
	cells := make([]sectorCellStatus, total)

	uncorrectable := sectorIssueCellCount(drive.UncorrectableSectors, total, 4)
	pending := sectorIssueCellCount(drive.PendingSectors, total, 3)
	reallocated := sectorIssueCellCount(drive.ReallocatedSectors, total, 3)

	if drive.HealthStatus == smart.HealthBad && uncorrectable == 0 {
		uncorrectable = maxInt(1, total/10)
	}
	if drive.HealthStatus == smart.HealthCaution && pending == 0 && reallocated == 0 && uncorrectable == 0 {
		pending = maxInt(1, total/16)
	}

	fillSectorCells(cells, sectorCellUncorrectable, uncorrectable, 3)
	fillSectorCells(cells, sectorCellPending, pending, 5)
	fillSectorCells(cells, sectorCellReallocated, reallocated, 7)
	return cells
}

func sectorIssueCellCount(count int64, total, maxDivisor int) int {
	if count <= 0 || total <= 0 {
		return 0
	}

	cells := int(count)
	if count > 12 {
		cells = 12 + int(count/12)
	}
	maxCells := maxInt(1, total/maxDivisor)
	if cells > maxCells {
		cells = maxCells
	}
	if cells < 1 {
		cells = 1
	}
	return cells
}

func fillSectorCells(cells []sectorCellStatus, status sectorCellStatus, count, seed int) {
	if count <= 0 || len(cells) == 0 {
		return
	}

	step := seed*2 + 1
	idx := (seed * 11) % len(cells)
	filled := 0
	attempts := 0
	for filled < count && attempts < len(cells)*3 {
		if cells[idx] == sectorCellGood {
			cells[idx] = status
			filled++
		}
		idx = (idx + step) % len(cells)
		attempts++
	}
}

func sectorGridDriveLabel(drive smart.DriveInfo) string {
	name := drive.Model
	if name == "" {
		name = drive.Device
	}
	if name == "" {
		name = "Unknown drive"
	}
	if drive.Device == "" || drive.Device == name {
		return name
	}
	return fmt.Sprintf("%s %s", name, drive.Device)
}

func sectorRiskScore(drive smart.DriveInfo) int {
	score := int(drive.ReallocatedSectors*3 + drive.PendingSectors*8 + drive.UncorrectableSectors*12 + drive.CRCErrors)
	switch drive.HealthStatus {
	case smart.HealthBad:
		score += 50
	case smart.HealthCaution:
		score += 20
	case smart.HealthInfo:
		score += 8
	}
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

func sectorGridStatus(drive smart.DriveInfo) smart.HealthStatus {
	hasSectorPressure := drive.ReallocatedSectors > 0 ||
		drive.PendingSectors > 0 ||
		drive.UncorrectableSectors > 0

	hasSevereSectorPressure := drive.ReallocatedSectors > smart.ReallocatedSectorsBadThreshold ||
		drive.PendingSectors > sectorGridPendingBadLimit ||
		drive.UncorrectableSectors > smart.UncorrectableSectorsBadThreshold

	switch {
	case drive.HealthStatus == smart.HealthBad || hasSevereSectorPressure:
		return smart.HealthBad
	case hasSectorPressure || drive.HealthStatus == smart.HealthCaution:
		return smart.HealthCaution
	case drive.CRCErrors > 0 || drive.HealthStatus == smart.HealthInfo:
		return smart.HealthInfo
	default:
		return smart.HealthGood
	}
}

func fallbackValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
