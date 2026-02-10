package screenshot

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// captureWindow is the function used to capture the active window (mockable for tests)
var captureWindow func(string) error = CaptureActiveWindow

// copyToClipboard is the function used to copy a file to clipboard (mockable for tests)
var copyToClipboard func(string) error = CopyFileToClipboard

// CaptureToClipboard captures the terminal window and copies it to clipboard
func CaptureToClipboard() error {
	// Create temp file path
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("nuther_screenshot_%s.png", timestamp))

	// Capture the active window
	if err := captureWindow(tempFile); err != nil {
		return fmt.Errorf("failed to capture window: %w", err)
	}

	// Copy image file to clipboard
	if err := copyToClipboard(tempFile); err != nil {
		// Try to clean up temp file
		if removeErr := os.Remove(tempFile); removeErr != nil {
			slog.Warn("failed to cleanup temp file", "path", tempFile, "error", removeErr)
		}
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	// Clean up temp file
	if err := os.Remove(tempFile); err != nil {
		slog.Warn("failed to cleanup temp file", "path", tempFile, "error", err)
	}

	return nil
}

// GetScreenshotPath returns a path for saving screenshots
func GetScreenshotPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	timestamp := time.Now().Format("20060102_150405")
	return filepath.Join(homeDir, fmt.Sprintf("nuther_screenshot_%s.png", timestamp))
}
