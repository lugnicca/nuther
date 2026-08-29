package ui

import (
	"nuther/internal/config"
	"nuther/internal/screenshot"
	"nuther/internal/smart"
	"nuther/internal/ui/styles"
	"nuther/internal/ui/views"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadDrivesCmd returns a command that loads all drives
func LoadDrivesCmd() tea.Cmd {
	return func() tea.Msg {
		drives, err := smart.ScanDrives()
		return DrivesLoadedMsg{
			Drives: drives,
			Error:  err,
		}
	}
}

// CaptureScreenshotCmd captures the terminal window and copies to clipboard
func CaptureScreenshotCmd() tea.Cmd {
	return func() tea.Msg {
		err := screenshot.CaptureToClipboard()
		if err != nil {
			return ScreenshotMsg{Success: false, Error: err}
		}
		return ScreenshotMsg{Success: true}
	}
}

func SaveOverviewImageCmd(drive smart.DriveInfo, selectedAttr, scrollOffset, width, height int, s *styles.Styles, cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		path := screenshot.GetScreenshotPath(cfg.Screenshot.Dir, drive.Device, drive.Serial)
		overview := views.RenderOverview(drive, selectedAttr, scrollOffset, width, height, s)
		if err := screenshot.RenderOverviewImage(overview, path, cfg); err != nil {
			return ScreenshotMsg{Success: false, Error: err}
		}
		return ScreenshotMsg{Success: true, Path: path}
	}
}
