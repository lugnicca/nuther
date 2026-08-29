package screenshot

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

// GetScreenshotPath returns a path for saving a drive overview image. The
// directory defaults to the user's home dir (or temp dir) when empty. The id
// is the drive's device identifier (e.g. /dev/disk3); only its basename is
// used to keep the filename short and space-free.
func GetScreenshotPath(dir, id, serial string) string {
	if dir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dir = home
		} else {
			dir = os.TempDir()
		}
	}

	id = strings.TrimSpace(id)
	if id == "" {
		id = "unknown"
	} else {
		id = filepath.Base(id)
	}
	id = sanitizeFilenamePart(id)

	serial = sanitizeFilenamePart(serial)
	if serial == "" {
		serial = "unknown"
	}

	timestamp := time.Now().Format("20060102_150405")
	return filepath.Join(dir, fmt.Sprintf("nuther_%s_%s_%s.png", id, serial, timestamp))
}

// sanitizeFilenamePart replaces characters unsafe in filenames (including
// spaces) with underscores.
func sanitizeFilenamePart(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', ' ':
			return '_'
		}
		if r < 0x20 {
			return '_'
		}
		return r
	}, strings.TrimSpace(s))
}
