//go:build unix

package smartwatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// With a corrupt index and a snapshots directory that cannot be searched, the
// store must return the index parse error at once (as v0.6.0 did) instead of
// reaching the uniqueness probe and looping there.
func TestStoreCorruptIndexWithUnsearchableSnapshotsFailsFast(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	dir := t.TempDir()
	snapshots := filepath.Join(dir, "snapshots")
	if err := os.Mkdir(snapshots, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chmod(snapshots, 0o000); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(snapshots, 0o755) })

	err := saveSnapshotWithTimeout(t, dir)
	if err == nil || !strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Fatalf("SaveSnapshot() error = %v, want the corrupt-index error", err)
	}
}

// Even with a readable index, a snapshots directory that cannot be searched
// must make the uniqueness probe give up with an error rather than loop.
func TestStoreUnsearchableSnapshotsReturnsError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	dir := t.TempDir()
	snapshots := filepath.Join(dir, "snapshots")
	if err := os.Mkdir(snapshots, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.Chmod(snapshots, 0o000); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(snapshots, 0o755) })

	err := saveSnapshotWithTimeout(t, dir)
	if err == nil {
		t.Fatal("SaveSnapshot() error = nil, want the uniqueness probe to give up")
	}
}

// saveSnapshotWithTimeout fails the test if SaveSnapshot does not return.
func saveSnapshotWithTimeout(t *testing.T, dir string) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		_, err := NewStore(dir).SaveSnapshot(time.Now(), ReasonManual, testDrive("/dev/sda", "SN-A"))
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("SaveSnapshot() hung")
		return nil
	}
}
