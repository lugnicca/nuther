//go:build windows

package screenshot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CaptureActiveWindow captures the active window to a file using PowerShell
func CaptureActiveWindow(outputPath string) error {
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}

	// PowerShell script to capture the active window.
	// The output path is passed via environment variable to prevent injection.
	psScript := `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

# Get the foreground window handle
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class Win32 {
    [DllImport("user32.dll")]
    public static extern IntPtr GetForegroundWindow();
    [DllImport("user32.dll")]
    public static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);
    [StructLayout(LayoutKind.Sequential)]
    public struct RECT {
        public int Left;
        public int Top;
        public int Right;
        public int Bottom;
    }
}
"@

Start-Sleep -Milliseconds 100

$hwnd = [Win32]::GetForegroundWindow()
$rect = New-Object Win32+RECT
[Win32]::GetWindowRect($hwnd, [ref]$rect) | Out-Null

$width = $rect.Right - $rect.Left
$height = $rect.Bottom - $rect.Top

$bitmap = New-Object System.Drawing.Bitmap($width, $height)
$graphics = [System.Drawing.Graphics]::FromImage($bitmap)
$graphics.CopyFromScreen($rect.Left, $rect.Top, 0, 0, $bitmap.Size)
$bitmap.Save($env:NUTHER_SCREENSHOT_PATH)
$graphics.Dispose()
$bitmap.Dispose()
`

	ctx, cancel := context.WithTimeout(context.Background(), ScreenshotTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psScript)
	cmd.Env = append(os.Environ(), "NUTHER_SCREENSHOT_PATH="+absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell error: %w, output: %s", err, string(output))
	}

	// Verify file was created
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("screenshot file was not created")
	}

	return nil
}

// CopyFileToClipboard copies an image file to the Windows clipboard
func CopyFileToClipboard(imagePath string) error {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return err
	}

	// PowerShell script to copy image to clipboard.
	// The image path is passed via environment variable to prevent injection.
	psScript := `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$image = [System.Drawing.Image]::FromFile($env:NUTHER_CLIPBOARD_PATH)
[System.Windows.Forms.Clipboard]::SetImage($image)
$image.Dispose()
`

	ctx, cancel := context.WithTimeout(context.Background(), ScreenshotTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psScript)
	cmd.Env = append(os.Environ(), "NUTHER_CLIPBOARD_PATH="+absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell error: %w, output: %s", err, string(output))
	}

	return nil
}
