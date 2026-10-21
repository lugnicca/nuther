//go:build windows

package smartwatch

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockReportFile takes a blocking exclusive lock on path+".lock". The lock
// lives on a stable sidecar rather than the report file itself because
// rotation replaces the report file's inode. It blocks until the lock is
// available, so a held lock waits rather than drops the report.
func lockReportFile(path string) (func() error, error) {
	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	overlapped := new(windows.Overlapped)
	handle := windows.Handle(f.Fd())
	if err := windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, overlapped); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() error {
		_ = windows.UnlockFileEx(handle, 0, 1, 0, overlapped)
		return f.Close()
	}, nil
}
