package screenshot

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestRenderTextImagePNGAndJPG(t *testing.T) {
	dir := t.TempDir()
	for _, ext := range []string{".png", ".jpg"} {
		path := filepath.Join(dir, "overview"+ext)
		if err := RenderTextImage("\x1b[31mOverview\x1b[0m\nSamsung SSD", path); err != nil {
			t.Fatalf("RenderTextImage(%s): %v", ext, err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open rendered image: %v", err)
		}
		cfg, _, err := image.DecodeConfig(f)
		f.Close()
		if err != nil {
			t.Fatalf("decode rendered image: %v", err)
		}
		if cfg.Width <= 0 || cfg.Height <= 0 {
			t.Fatalf("invalid image size: %dx%d", cfg.Width, cfg.Height)
		}
	}
}
