//go:build !windows && !darwin && !linux

package screenshot

// CaptureActiveWindow is not supported on this platform
func CaptureActiveWindow(outputPath string) error {
	return ErrUnsupportedPlatform
}

// CopyFileToClipboard is not supported on this platform
func CopyFileToClipboard(imagePath string) error {
	return ErrUnsupportedPlatform
}
