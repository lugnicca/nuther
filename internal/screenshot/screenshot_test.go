package screenshot

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetScreenshotPath(t *testing.T) {
	path := GetScreenshotPath("", "/dev/disk3", "S64ANF0R123456")

	if path == "" {
		t.Error("GetScreenshotPath should return a non-empty path")
	}

	// Should end with .png
	if !strings.HasSuffix(path, ".png") {
		t.Errorf("Path should end with .png, got %q", path)
	}

	// Should contain the device basename and serial, with no spaces
	filename := filepath.Base(path)
	if strings.Contains(filename, " ") {
		t.Errorf("Filename should not contain spaces, got %q", filename)
	}
	if !strings.Contains(filename, "disk3") {
		t.Errorf("Filename should contain device basename, got %q", filename)
	}
	if !strings.Contains(filename, "S64ANF0R123456") {
		t.Errorf("Filename should contain serial, got %q", filename)
	}

	// Should be in home dir or temp dir
	homeDir, _ := os.UserHomeDir()
	tempDir := os.TempDir()
	if !strings.HasPrefix(path, homeDir) && !strings.HasPrefix(path, tempDir) {
		t.Errorf("Path should be in home or temp dir, got %q", path)
	}
}

func TestGetScreenshotPathContainsTimestamp(t *testing.T) {
	path1 := GetScreenshotPath("", "nvme0n1", "Serial")
	// Wait a tiny bit to ensure different timestamp
	path2 := GetScreenshotPath("", "nvme0n1", "Serial")

	// Paths should include timestamp format YYYYMMDD_HHMMSS
	// They might be the same if called within the same second
	if path1 == "" || path2 == "" {
		t.Error("Paths should not be empty")
	}

	// Both should have the expected filename pattern
	filename1 := filepath.Base(path1)
	if !strings.HasPrefix(filename1, "nuther_nvme0n1_Serial_") {
		t.Errorf("Filename should start with nuther_nvme0n1_Serial_, got %q", filename1)
	}
}

func TestGetScreenshotPathDirectory(t *testing.T) {
	dir := t.TempDir()
	path := GetScreenshotPath(dir, "disk3", "Serial")

	// Directory should be the configured dir
	if got := filepath.Dir(path); got != dir {
		t.Errorf("Directory = %q, want %q", got, dir)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Directory %q should exist: %v", dir, err)
	}
	if !info.IsDir() {
		t.Errorf("%q should be a directory", dir)
	}
}

func TestGetScreenshotPathSanitizesUnsafeCharacters(t *testing.T) {
	path := GetScreenshotPath(t.TempDir(), `Samsung SSD 980 PRO:With?`, `S<e>r*i"a|l`)
	filename := filepath.Base(path)

	if strings.ContainsAny(filename, `\/:*?"<>| `) {
		t.Errorf("Filename should not contain unsafe characters or spaces, got %q", filename)
	}
	if !strings.Contains(filename, "Samsung_SSD_980_PRO_With") {
		t.Errorf("Filename should contain sanitized id, got %q", filename)
	}
	if !strings.Contains(filename, "S_e_r_i_a_l") {
		t.Errorf("Filename should contain sanitized serial, got %q", filename)
	}
}

func TestGetScreenshotPathUsesDeviceBasename(t *testing.T) {
	path := GetScreenshotPath(t.TempDir(), "/dev/disk3", "S64ANF0R123456")
	filename := filepath.Base(path)

	if !strings.Contains(filename, "nuther_disk3_S64ANF0R123456_") {
		t.Errorf("Filename should use the device basename, got %q", filename)
	}
	if strings.Contains(filename, " ") {
		t.Errorf("Filename should not contain spaces, got %q", filename)
	}
}

func TestGetScreenshotPathModelLikeIdNoSpaces(t *testing.T) {
	path := GetScreenshotPath(t.TempDir(), "Samsung SSD 980 PRO", "S64ANF0R123456")
	filename := filepath.Base(path)

	if strings.Contains(filename, " ") {
		t.Errorf("Filename should not contain spaces, got %q", filename)
	}
	if !strings.Contains(filename, "Samsung_SSD_980_PRO") {
		t.Errorf("Filename should contain sanitized id, got %q", filename)
	}
}

func TestGetScreenshotPathFallsBackToUnknown(t *testing.T) {
	path := GetScreenshotPath(t.TempDir(), "", "")
	filename := filepath.Base(path)

	if !strings.Contains(filename, "nuther_unknown_unknown_") {
		t.Errorf("Filename should use unknown fallbacks, got %q", filename)
	}
}

func TestErrorTypes(t *testing.T) {
	// Verify error types are defined correctly
	if ErrUnsupportedPlatform == nil {
		t.Error("ErrUnsupportedPlatform should be defined")
	}

	if ErrClipboardFailed == nil {
		t.Error("ErrClipboardFailed should be defined")
	}

	// Check error messages
	if ErrUnsupportedPlatform.Error() == "" {
		t.Error("ErrUnsupportedPlatform should have a message")
	}

	if ErrClipboardFailed.Error() == "" {
		t.Error("ErrClipboardFailed should have a message")
	}
}

func TestScreenshotTimeout(t *testing.T) {
	if ScreenshotTimeout <= 0 {
		t.Errorf("ScreenshotTimeout should be positive, got %v", ScreenshotTimeout)
	}
}

// withMockedCapture sets mock functions and restores originals after test
func withMockedCapture(t *testing.T, captureFn func(string) error, clipboardFn func(string) error) {
	t.Helper()
	origCapture := captureWindow
	origClipboard := copyToClipboard
	captureWindow = captureFn
	copyToClipboard = clipboardFn
	t.Cleanup(func() {
		captureWindow = origCapture
		copyToClipboard = origClipboard
	})
}

func TestCaptureToClipboard_Success(t *testing.T) {
	// Mock: capture creates a temp file, clipboard succeeds
	withMockedCapture(t,
		func(path string) error {
			return os.WriteFile(path, []byte("fake png"), 0644)
		},
		func(path string) error {
			return nil
		},
	)

	err := CaptureToClipboard()
	if err != nil {
		t.Errorf("CaptureToClipboard() error = %v, want nil", err)
	}
}

func TestCaptureToClipboard_CaptureError(t *testing.T) {
	captureErr := errors.New("capture failed")
	withMockedCapture(t,
		func(path string) error {
			return captureErr
		},
		func(path string) error {
			return nil
		},
	)

	err := CaptureToClipboard()
	if err == nil {
		t.Error("CaptureToClipboard() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to capture window") {
		t.Errorf("error should mention capture failure, got %q", err.Error())
	}
}

func TestCaptureToClipboard_ClipboardError(t *testing.T) {
	clipErr := errors.New("clipboard failed")
	withMockedCapture(t,
		func(path string) error {
			return os.WriteFile(path, []byte("fake png"), 0644)
		},
		func(path string) error {
			return clipErr
		},
	)

	err := CaptureToClipboard()
	if err == nil {
		t.Error("CaptureToClipboard() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to copy to clipboard") {
		t.Errorf("error should mention clipboard failure, got %q", err.Error())
	}
}

func TestSaveScreenshot_Success(t *testing.T) {
	withMockedCapture(t,
		func(path string) error {
			return os.WriteFile(path, []byte("fake png"), 0644)
		},
		func(path string) error {
			return nil
		},
	)

	path, err := SaveScreenshot()
	if err != nil {
		t.Errorf("SaveScreenshot() error = %v, want nil", err)
	}
	if path == "" {
		t.Error("SaveScreenshot() returned empty path")
	}
	if !strings.HasSuffix(path, ".png") {
		t.Errorf("SaveScreenshot() path should end with .png, got %q", path)
	}

	// Clean up the created file
	os.Remove(path)
}

func TestSaveScreenshot_Error(t *testing.T) {
	captureErr := errors.New("capture failed")
	withMockedCapture(t,
		func(path string) error {
			return captureErr
		},
		func(path string) error {
			return nil
		},
	)

	path, err := SaveScreenshot()
	if err == nil {
		t.Error("SaveScreenshot() error = nil, want error")
	}
	if path != "" {
		t.Errorf("SaveScreenshot() path = %q, want empty on error", path)
	}
}

func TestCopyFileToClipboard_NonExistent(t *testing.T) {
	err := CopyFileToClipboard("/nonexistent/path/file.png")
	// This should fail because the file doesn't exist
	// The exact error depends on platform, but it should not panic
	_ = err
}

func TestCaptureActiveWindow_InvalidPath(t *testing.T) {
	// Call CaptureActiveWindow with a temp path - it will try to capture
	// and may fail (no GUI in test env), but should not panic
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_capture.png")

	err := CaptureActiveWindow(outputPath)
	// We don't assert success since it depends on GUI availability
	// We just verify no panic and the function returns
	_ = err
}

func TestCaptureToClipboard_SuccessCleanup(t *testing.T) {
	// Test the success path where the temp file is cleaned up after clipboard copy
	var capturedPath string
	withMockedCapture(t,
		func(path string) error {
			capturedPath = path
			return os.WriteFile(path, []byte("fake png data"), 0644)
		},
		func(path string) error {
			return nil
		},
	)

	err := CaptureToClipboard()
	if err != nil {
		t.Errorf("CaptureToClipboard() error = %v, want nil", err)
	}

	// Temp file should have been cleaned up
	if capturedPath != "" {
		if _, statErr := os.Stat(capturedPath); !os.IsNotExist(statErr) {
			t.Error("Temp file should have been cleaned up after successful clipboard copy")
		}
	}
}

func TestCaptureToClipboard_ClipboardErrorCleansUp(t *testing.T) {
	// Test that temp file is cleaned up even when clipboard fails
	var capturedPath string
	withMockedCapture(t,
		func(path string) error {
			capturedPath = path
			return os.WriteFile(path, []byte("fake png data"), 0644)
		},
		func(path string) error {
			return errors.New("clipboard unavailable")
		},
	)

	err := CaptureToClipboard()
	if err == nil {
		t.Error("CaptureToClipboard() should return error")
	}

	// Temp file should have been cleaned up despite error
	if capturedPath != "" {
		if _, statErr := os.Stat(capturedPath); !os.IsNotExist(statErr) {
			t.Error("Temp file should be cleaned up even when clipboard fails")
		}
	}
}

func TestCopyFileToClipboard_ExistingFile(t *testing.T) {
	// Create a real temp file and attempt clipboard copy
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(tmpFile, []byte("fake png"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err := CopyFileToClipboard(tmpFile)
	// May fail (no clipboard in test env) but should not panic
	_ = err
}

