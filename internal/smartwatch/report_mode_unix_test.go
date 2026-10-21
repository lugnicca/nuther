//go:build unix

package smartwatch

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// The 0644 mode passed to OpenFile is masked by umask, and rotation replaces
// the file, so the tool must set the mode explicitly: under umask 0077 a
// non-root tailer such as Alloy would otherwise lose read access while the
// scan still reports success.
func TestAppendReportModeSurvivesRotationUnderRestrictiveUmask(t *testing.T) {
	previous := syscall.Umask(0o077)
	defer syscall.Umask(previous)

	path := filepath.Join(t.TempDir(), "report.jsonl")
	snapshot := reportSnapshotAt("/dev/sda", "SN-A", time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC))

	if err := AppendReport(path, []Snapshot{snapshot}); err != nil {
		t.Fatalf("AppendReport() error = %v", err)
	}
	assertReportMode(t, path, "after create")

	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), ReportLogMaxBytes), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := AppendReport(path, []Snapshot{snapshot}); err != nil {
		t.Fatalf("AppendReport() after seeding error = %v", err)
	}
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("rotation did not happen: %v", err)
	}
	assertReportMode(t, path, "after rotation")
}

func assertReportMode(t *testing.T, path, when string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Fatalf("report mode %s = %o, want 644", when, got)
	}
}
