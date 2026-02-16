package styles

import (
	"fmt"

	"nuther/internal/config"
	"nuther/internal/smart"

	"github.com/charmbracelet/lipgloss"
)

// Icons for the UI - minimalist Unicode symbols
const (
	IconSuccess    = "✓"
	IconWarning    = "!"
	IconError      = "✗"
	IconInfo       = "·"
	IconBullet     = "●"
	IconArrowRight = "→"
	IconHDD        = "◇"
	IconSSD        = "◆"
	IconNVMe       = "▪"
	IconUSB        = "○"
	IconHealth     = "+"
	IconTemp       = "°"
	IconClock      = "◷"
	IconPower      = "⏻"
	IconPlug       = "▸"
	IconGear       = "∗"
	IconWave       = "~"
)

// Box drawing characters
const (
	BoxTopLeft     = "╭"
	BoxTopRight    = "╮"
	BoxBottomLeft  = "╰"
	BoxBottomRight = "╯"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	ProgressFull   = "█"
	ProgressEmpty  = "░"
)

// Styles contains all application styles
type Styles struct {
	// Display settings
	ShowFahrenheit bool

	// Colors
	AccentPrimary   lipgloss.Color
	AccentSecondary lipgloss.Color
	Success         lipgloss.Color
	Warning         lipgloss.Color
	Danger          lipgloss.Color
	Info            lipgloss.Color
	Text            lipgloss.Color
	TextDim         lipgloss.Color
	Background      lipgloss.Color
	Border          lipgloss.Color
	tempHot         lipgloss.Color // orange for hot temperature, not exported

	// Base styles
	Base      lipgloss.Style
	Bold      lipgloss.Style
	Italic    lipgloss.Style
	Dim       lipgloss.Style

	// Header
	Logo     lipgloss.Style
	Subtitle lipgloss.Style

	// Tabs
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style

	// Drive selector
	DriveButton         lipgloss.Style
	DriveButtonSelected lipgloss.Style

	// Health badges
	HealthGood    lipgloss.Style
	HealthInfo    lipgloss.Style
	HealthCaution lipgloss.Style
	HealthBad     lipgloss.Style

	// Table
	TableHeader      lipgloss.Style
	TableRow         lipgloss.Style
	TableRowAlt      lipgloss.Style
	TableRowSelected lipgloss.Style

	// Cards
	Card       lipgloss.Style
	CardTitle  lipgloss.Style
	CardValue  lipgloss.Style
	CardLabel  lipgloss.Style
	CardBorder lipgloss.Style

	// Status bar
	StatusBar     lipgloss.Style
	StatusBarItem lipgloss.Style

	// Help
	HelpKey     lipgloss.Style
	HelpDesc    lipgloss.Style
	HelpOverlay lipgloss.Style

	// Box
	Box      lipgloss.Style
	BoxTitle lipgloss.Style
}

// NewStyles creates styles from configuration
func NewStyles(cfg *config.Config) *Styles {
	s := &Styles{}

	// Display settings
	s.ShowFahrenheit = cfg.Display.ShowFahrenheit

	// Set colors from config
	s.AccentPrimary = lipgloss.Color(cfg.Colors.AccentPrimary)
	s.AccentSecondary = lipgloss.Color(cfg.Colors.AccentSecondary)
	s.Success = lipgloss.Color(cfg.Colors.Success)
	s.Warning = lipgloss.Color(cfg.Colors.Warning)
	s.Danger = lipgloss.Color(cfg.Colors.Danger)
	s.Info = lipgloss.Color(cfg.Colors.Info)
	s.Text = lipgloss.Color(cfg.Colors.Text)
	s.TextDim = lipgloss.Color(cfg.Colors.TextDim)
	s.Background = lipgloss.Color(cfg.Colors.Background)
	s.Border = lipgloss.Color(cfg.Colors.Border)

	badgeFg := lipgloss.Color(cfg.Colors.BadgeForeground)
	badgeFgAlt := lipgloss.Color(cfg.Colors.BadgeForegroundAlt)
	surfaceAlt := lipgloss.Color(cfg.Colors.SurfaceAlt)
	surfaceHighlight := lipgloss.Color(cfg.Colors.SurfaceHighlight)
	s.tempHot = lipgloss.Color(cfg.Colors.TempHot)

	// Base styles
	s.Base = lipgloss.NewStyle().Foreground(s.Text)
	s.Bold = lipgloss.NewStyle().Bold(true)
	s.Italic = lipgloss.NewStyle().Italic(true)
	s.Dim = lipgloss.NewStyle().Foreground(s.TextDim)

	// Logo style
	s.Logo = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.AccentPrimary)

	s.Subtitle = lipgloss.NewStyle().
		Italic(true).
		Foreground(s.TextDim)

	// Tab styles
	s.Tab = lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(s.TextDim)

	s.ActiveTab = lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true).
		Foreground(badgeFg).
		Background(s.AccentSecondary)

	// Drive button styles
	s.DriveButton = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(s.TextDim)

	s.DriveButtonSelected = lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(badgeFg).
		Background(s.AccentPrimary)

	// Health badges
	s.HealthGood = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(badgeFg).
		Background(s.Success)

	s.HealthInfo = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(badgeFg).
		Background(s.Info)

	s.HealthCaution = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(badgeFg).
		Background(s.Warning)

	s.HealthBad = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(badgeFgAlt).
		Background(s.Danger)

	// Table styles
	s.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.AccentPrimary)

	s.TableRow = lipgloss.NewStyle().
		Foreground(s.Text)

	s.TableRowAlt = lipgloss.NewStyle().
		Foreground(s.Text).
		Background(surfaceAlt)

	s.TableRowSelected = lipgloss.NewStyle().
		Foreground(s.Text).
		Background(surfaceHighlight)

	// Card styles
	s.Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Border).
		Padding(0, 1).
		Width(19)

	s.CardTitle = lipgloss.NewStyle().
		Foreground(s.TextDim)

	s.CardValue = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Text)

	s.CardLabel = lipgloss.NewStyle().
		Foreground(s.TextDim)

	s.CardBorder = lipgloss.NewStyle().
		Foreground(s.Border)

	// Status bar
	s.StatusBar = lipgloss.NewStyle().
		Foreground(s.TextDim).
		Background(surfaceAlt).
		Padding(0, 1)

	s.StatusBarItem = lipgloss.NewStyle().
		Padding(0, 1)

	// Help styles
	s.HelpKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.AccentPrimary)

	s.HelpDesc = lipgloss.NewStyle().
		Foreground(s.TextDim)

	s.HelpOverlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.AccentPrimary).
		Padding(1, 2).
		Background(s.Background)

	// Box styles
	s.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Border).
		Padding(0, 1)

	s.BoxTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.AccentPrimary)

	return s
}

// GetHealthStyle returns the appropriate style for a health status
func (s *Styles) GetHealthStyle(status smart.HealthStatus) lipgloss.Style {
	switch status {
	case smart.HealthGood:
		return s.HealthGood
	case smart.HealthInfo:
		return s.HealthInfo
	case smart.HealthCaution:
		return s.HealthCaution
	case smart.HealthBad:
		return s.HealthBad
	default:
		return s.Dim
	}
}

// GetHealthColor returns the appropriate color for a health status
func (s *Styles) GetHealthColor(status smart.HealthStatus) lipgloss.Color {
	switch status {
	case smart.HealthGood:
		return s.Success
	case smart.HealthInfo:
		return s.Info
	case smart.HealthCaution:
		return s.Warning
	case smart.HealthBad:
		return s.Danger
	default:
		return s.TextDim
	}
}

// GetHealthIcon returns the icon for a health status
func (s *Styles) GetHealthIcon(status smart.HealthStatus) string {
	switch status {
	case smart.HealthGood:
		return IconSuccess
	case smart.HealthInfo:
		return IconInfo
	case smart.HealthCaution:
		return IconWarning
	case smart.HealthBad:
		return IconError
	default:
		return IconInfo
	}
}

// GetDriveIcon returns the icon for a drive type
func (s *Styles) GetDriveIcon(drive *smart.DriveInfo) string {
	if drive.IsNVMe {
		return IconNVMe
	}
	if drive.IsSSD {
		return IconSSD
	}
	if drive.IsUSB {
		return IconUSB
	}
	return IconHDD
}

// GetTemperatureColor returns the color for a temperature value
func (s *Styles) GetTemperatureColor(temp int) lipgloss.Color {
	switch {
	case temp < smart.TempCool:
		return s.Info
	case temp < smart.TempGoodMax:
		return s.Success
	case temp < smart.TempCautionMax:
		return s.Warning
	case temp < smart.TempCritical:
		return s.tempHot
	default:
		return s.Danger
	}
}

// FormatTemperature formats temperature based on user preference (Celsius or Fahrenheit)
func (s *Styles) FormatTemperature(temp int) string {
	if s.ShowFahrenheit {
		fahrenheit := (temp * 9 / 5) + 32
		return fmt.Sprintf("%d°F", fahrenheit)
	}
	return fmt.Sprintf("%d°C", temp)
}

