//go:build !windows

package smartwatch

import (
	"os"

	"golang.org/x/sys/unix"
)

// lockReportFile takes a blocking exclusive flock on path+".lock". The lock
// lives on a stable sidecar rather than the report file itself because
// rotation replaces the report file's inode. It blocks until the lock is
// available, so a held lock waits rather than drops the report.
func lockReportFile(path string) (func() error, error) {
	f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() error {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		return f.Close()
	}, nil
}
