package ui

import (
	"nuther/internal/smart"
	"time"
)

// DrivesLoadedMsg is sent when drives are loaded
type DrivesLoadedMsg struct {
	Drives []smart.DriveInfo
	Error  error
}

// DriveRefreshedMsg is sent when a single drive is refreshed
type DriveRefreshedMsg struct {
	Index int
	Drive smart.DriveInfo
	Error error
}

// ScreenshotMsg is sent when screenshot is complete
type ScreenshotMsg struct {
	Success bool
	Error   error
	Path    string // Path to saved screenshot file (if saved to file)
}

// TickMsg for periodic updates
type TickMsg time.Time
