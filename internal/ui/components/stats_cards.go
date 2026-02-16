package components

import (
	"fmt"
	"strings"

	"nuther/internal/smart"
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatsCards renders the stats cards (temperature, power-on hours, power cycles, data written)
func RenderStatsCards(drive smart.DriveInfo, s *styles.Styles) string {
	var cards strings.Builder

	tempColor := s.GetTemperatureColor(drive.Temperature)
	borderStyle := lipgloss.NewStyle().Foreground(s.Border)

	// Box width: 17 chars inside + 2 for borders = 19 total
	// But emojis take 2 display columns, so we need 16 chars after emoji
	const boxWidth = 17

	// Stats row with boxes
	cards.WriteString(" ")

	// Top borders
	cards.WriteString(borderStyle.Render("╭─────────────────╮"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╭─────────────────╮"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╭─────────────────╮"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╭─────────────────╮"))
	cards.WriteString("\n")

	// Labels row - each label must be exactly 17 display columns
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(s.Dim.Render(PadRight(" TEMPERATURE", boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(s.Dim.Render(PadRight(" POWER ON HRS", boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(s.Dim.Render(PadRight(" PWR CYCLES", boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(s.Dim.Render(PadRight(" DATA WRITTEN", boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString("\n")

	// Values row - each value must be exactly 17 display columns
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	tempStr := fmt.Sprintf(" %s", s.FormatTemperature(drive.Temperature))
	cards.WriteString(lipgloss.NewStyle().Bold(true).Foreground(tempColor).Render(PadRight(tempStr, boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	hoursStr := fmt.Sprintf(" %s", smart.FormatNumber(drive.PowerOnHours))
	cards.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.AccentPrimary).Render(PadRight(hoursStr, boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	cyclesStr := fmt.Sprintf(" %s", smart.FormatNumber(drive.PowerCycles))
	cards.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.AccentSecondary).Render(PadRight(cyclesStr, boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("│"))
	writtenStr := " —"
	if drive.TotalBytesWritten >= 0 {
		writtenStr = " " + smart.FormatBytes(drive.TotalBytesWritten)
	}
	cards.WriteString(lipgloss.NewStyle().Bold(true).Foreground(s.Info).Render(PadRight(writtenStr, boxWidth)))
	cards.WriteString(borderStyle.Render("│"))
	cards.WriteString("\n")

	// Bottom borders
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╰─────────────────╯"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╰─────────────────╯"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╰─────────────────╯"))
	cards.WriteString(" ")
	cards.WriteString(borderStyle.Render("╰─────────────────╯"))
	cards.WriteString("\n")

	return cards.String()
}
