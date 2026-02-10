package components

import (
	"nuther/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderHelp renders the help overlay
func RenderHelp(s *styles.Styles) string {
	keyStyle := s.HelpKey
	descStyle := s.HelpDesc

	help := `
  Navigation:
    ` + keyStyle.Render("↑/k, ↓/j") + `     ` + descStyle.Render("Navigate / scroll") + `
    ` + keyStyle.Render("←/h, →/l") + `     ` + descStyle.Render("Change option (settings)") + `
    ` + keyStyle.Render("n/], p/[") + `     ` + descStyle.Render("Next/previous drive") + `
    ` + keyStyle.Render("Tab") + `          ` + descStyle.Render("Switch tab") + `

  Actions:
    ` + keyStyle.Render("r") + `            ` + descStyle.Render("Refresh data") + `
    ` + keyStyle.Render("s") + `            ` + descStyle.Render("Screenshot to clipboard") + `
    ` + keyStyle.Render("Enter") + `        ` + descStyle.Render("Apply (settings)") + `
    ` + keyStyle.Render("?") + `            ` + descStyle.Render("Toggle help") + `
    ` + keyStyle.Render("q") + `            ` + descStyle.Render("Quit") + `

  Health Indicators:
    ` + lipgloss.NewStyle().Foreground(s.Success).Bold(true).Render(styles.IconBullet+" GOOD") + `       ` + descStyle.Render("Drive is healthy") + `
    ` + lipgloss.NewStyle().Foreground(s.Warning).Bold(true).Render(styles.IconBullet+" CAUTION") + `    ` + descStyle.Render("Monitoring recommended") + `
    ` + lipgloss.NewStyle().Foreground(s.Danger).Bold(true).Render(styles.IconBullet+" BAD") + `        ` + descStyle.Render("Action required") + `
`

	return s.HelpOverlay.Width(52).Render(help)
}
