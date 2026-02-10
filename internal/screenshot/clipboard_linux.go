//go:build linux

package screenshot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// CaptureActiveWindow captures the active window to a file
func CaptureActiveWindow(outputPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ScreenshotTimeout)
	defer cancel()

	// Try different screenshot tools in order of preference

	// 1. Try gnome-screenshot (GNOME)
	if _, err := exec.LookPath("gnome-screenshot"); err == nil {
		cmd := exec.CommandContext(ctx, "gnome-screenshot", "-w", "-f", outputPath)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// 2. Try maim with slop (X11)
	if _, err := exec.LookPath("maim"); err == nil {
		// Get active window ID
		xdoCmd := exec.CommandContext(ctx, "xdotool", "getactivewindow")
		windowID, err := xdoCmd.Output()
		if err == nil {
			cmd := exec.CommandContext(ctx, "maim", "-i", string(windowID[:len(windowID)-1]), outputPath)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		// Fallback: capture with selection
		cmd := exec.CommandContext(ctx, "maim", "-s", outputPath)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// 3. Try scrot (X11)
	if _, err := exec.LookPath("scrot"); err == nil {
		cmd := exec.CommandContext(ctx, "scrot", "-u", outputPath)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// 4. Try import from ImageMagick (X11)
	if _, err := exec.LookPath("import"); err == nil {
		cmd := exec.CommandContext(ctx, "import", "-window", "root", outputPath)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// 5. Try grim (Wayland)
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if _, err := exec.LookPath("grim"); err == nil {
			// Try with slurp for window selection
			if _, err := exec.LookPath("slurp"); err == nil {
				slurpCmd := exec.CommandContext(ctx, "slurp")
				geometry, err := slurpCmd.Output()
				if err == nil {
					cmd := exec.CommandContext(ctx, "grim", "-g", string(geometry[:len(geometry)-1]), outputPath)
					if err := cmd.Run(); err == nil {
						return nil
					}
				}
			}
			// Fallback: full screen
			cmd := exec.CommandContext(ctx, "grim", outputPath)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("no screenshot tool available. Install one of: gnome-screenshot, maim, scrot, import (ImageMagick), or grim (Wayland)")
}

// CopyFileToClipboard copies an image file to the clipboard
func CopyFileToClipboard(imagePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ScreenshotTimeout)
	defer cancel()

	// Read the image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return err
	}

	// Check for Wayland
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd := exec.CommandContext(ctx, "wl-copy", "--type", "image/png")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return err
			}
			if err := cmd.Start(); err != nil {
				return err
			}
			if _, err := stdin.Write(imageData); err != nil {
				stdin.Close()
				return err
			}
			stdin.Close()
			return cmd.Wait()
		}
	}

	// Try xclip (X11)
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", "image/png", "-i", imagePath)
		return cmd.Run()
	}

	// Try xsel (X11) - note: xsel doesn't support images well
	return fmt.Errorf("no clipboard tool available. Install xclip or wl-copy")
}
