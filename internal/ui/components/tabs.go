package components

import (
	"strings"

	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderTabs renders the tab bar
func RenderTabs(tabs []string, activeTab, width int, s *styles.Styles) string {
	var renderedTabs []string

	for i, tab := range tabs {
		var tabStyle lipgloss.Style
		if i == activeTab {
			tabStyle = s.ActiveTab
		} else {
			tabStyle = s.Tab
		}
		renderedTabs = append(renderedTabs, tabStyle.Render(tab))
	}

	tabRow := strings.Join(renderedTabs, " ")
	separator := RenderHorizontalLine(width-2, s)

	return tabRow + "\n" + separator
}

// RenderHorizontalLine renders a horizontal line
func RenderHorizontalLine(width int, s *styles.Styles) string {
	line := strings.Repeat(styles.BoxHorizontal, width)
	return lipgloss.NewStyle().Foreground(s.Border).Render(line)
}
