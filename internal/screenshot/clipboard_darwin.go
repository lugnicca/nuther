//go:build darwin

package screenshot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CaptureActiveWindow captures the screen to a file using screencapture
func CaptureActiveWindow(outputPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ScreenshotTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "screencapture", "-x", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("screencapture failed: %w", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("screenshot file was not created")
	}

	return nil
}

// CopyFileToClipboard copies an image file to the macOS clipboard using osascript
func CopyFileToClipboard(imagePath string) error {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("image file does not exist: %s", absPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ScreenshotTimeout)
	defer cancel()

	script := fmt.Sprintf(
		`set the clipboard to (read (POSIX file %q) as «class PNGf»)`, absPath)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript clipboard error: %w, output: %s", err, string(output))
	}

	return nil
}
