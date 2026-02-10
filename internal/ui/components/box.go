package components

import (
	"strings"

	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// RenderBox draws a box around content with an optional title
func RenderBox(content string, width int, title string, s *styles.Styles) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Border).
		Width(width)

	if title != "" {
		// Add title styling
		titleStyled := s.BoxTitle.Render(title)
		content = titleStyled + "\n" + content
	}

	return boxStyle.Render(content)
}

// RenderGauge renders a horizontal gauge/progress bar
func RenderGauge(value, max, width int, color lipgloss.Color, s *styles.Styles) string {
	if max == 0 {
		max = 100
	}

	filled := int(float64(value) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	var gauge strings.Builder

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(s.TextDim)

	for i := 0; i < width; i++ {
		if i < filled {
			gauge.WriteString(filledStyle.Render(styles.ProgressFull))
		} else {
			gauge.WriteString(emptyStyle.Render(styles.ProgressEmpty))
		}
	}

	return gauge.String()
}

// PadRight pads a string to the specified display width with spaces on the right
// Takes into account wide characters like emojis that occupy 2 columns
func PadRight(str string, width int) string {
	strWidth := runewidth.StringWidth(str)
	if strWidth >= width {
		return runewidth.Truncate(str, width, "")
	}
	return str + strings.Repeat(" ", width-strWidth)
}
