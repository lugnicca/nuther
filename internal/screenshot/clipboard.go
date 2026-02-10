package screenshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ScreenshotTimeout is the maximum time to wait for screenshot subprocess commands
const ScreenshotTimeout = 10 * time.Second

// ErrUnsupportedPlatform is returned when the platform is not supported
var ErrUnsupportedPlatform = errors.New("unsupported platform for clipboard operation")

// ErrClipboardFailed is returned when clipboard operation fails
var ErrClipboardFailed = errors.New("clipboard operation failed")

// SaveScreenshot saves a screenshot to the user's home directory
func SaveScreenshot() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(homeDir, fmt.Sprintf("nuther_screenshot_%s.png", timestamp))

	if err := captureWindow(outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}
