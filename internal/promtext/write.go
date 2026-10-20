package promtext

import (
	"os"
	"path/filepath"
	"time"

	"nuther/internal/smart"
)

// WriteFile renders drives with now and atomically replaces dir/FileName.
// The content is written to a temp file in dir, chmod 0644, then renamed over
// the target. On any error the existing file is left untouched.
func WriteFile(dir string, drives []smart.DriveInfo, now time.Time) error {
	return writeAtomic(filepath.Join(dir, FileName), func() ([]byte, error) {
		return []byte(Render(drives, now)), nil
	})
}

// writeAtomic writes the bytes produced by render to path atomically. render is
// called before the temp file is created, so a render error leaves path as-is.
func writeAtomic(path string, render func() ([]byte, error)) error {
	content, err := render()
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".nuther-smart-*.prom")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
