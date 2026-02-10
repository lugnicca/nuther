package components

import (
	"fmt"
	"strings"
	"time"

	"nuther/internal/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMapInfo provides the keybinding information needed by the status bar
type KeyMapInfo interface {
	ShortHelp() []key.Binding
}

// RenderStatusBar renders the status bar at the bottom
func RenderStatusBar(selectedDrive, totalDrives int, lastRefresh time.Time, isDemo bool, cacheFresh bool, width int, keymap KeyMapInfo, s *styles.Styles) string {
	left := fmt.Sprintf(" Drive %d/%d", selectedDrive+1, totalDrives)
	if isDemo {
		left = " [DEMO]" + left
	}
	center := fmt.Sprintf("Updated: %s", lastRefresh.Format("15:04:05"))
	if cacheFresh {
		center += " (cached)"
	}

	// Build right text dynamically from keybindings
	var parts []string
	for _, b := range keymap.ShortHelp() {
		h := b.Help()
		parts = append(parts, h.Key+":"+h.Desc)
	}
	right := strings.Join(parts, "  ") + " "

	// Calculate padding
	leftLen := len(left)
	centerLen := len(center)
	rightLen := len(right)

	padding := width - leftLen - centerLen - rightLen
	if padding < 0 {
		padding = 0
	}

	leftPad := padding / 2
	rightPad := padding - leftPad

	content := left + strings.Repeat(" ", leftPad) + center + strings.Repeat(" ", rightPad) + right

	return s.StatusBar.Width(width).Render(content)
}

// RenderScreenshotStatus renders a notification for screenshot operations
func RenderScreenshotStatus(status, message string, s *styles.Styles) string {
	var icon string
	var color lipgloss.Color

	switch status {
	case "capturing":
		icon = "📷"
		color = s.Info
	case "success":
		icon = "✓"
		color = s.Success
	case "error":
		icon = "✗"
		color = s.Danger
	default:
		return ""
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Padding(0, 1)

	return "\n " + style.Render(icon+" "+message)
}
