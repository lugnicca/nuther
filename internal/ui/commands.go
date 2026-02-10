package ui

import (
	"nuther/internal/screenshot"
	"nuther/internal/smart"

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
