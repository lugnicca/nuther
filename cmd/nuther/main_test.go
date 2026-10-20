package main

import (
	"os"
	"path/filepath"
	"testing"

	"nuther/internal/smart"
)

// A --once run whose textfile publication fails must still exit non-zero and
// must still have saved the snapshot.
func TestRunSmartWatcherOnceFailsWhenPublicationFails(t *testing.T) {
	previous := scanDrives
	scanDrives = func() ([]smart.DriveInfo, error) {
		return []smart.DriveInfo{{
			Device:       "/dev/sda",
			Model:        "Test Drive",
			Serial:       "SN-A",
			Interface:    "SATA SSD",
			HealthStatus: smart.HealthGood,
			HealthPassed: true,
		}}, nil
	}
	defer func() { scanDrives = previous }()

	storeDir := t.TempDir()
	err := runSmartWatcher("watch-smart", []string{
		"--once",
		"--textfile", filepath.Join(t.TempDir(), "does-not-exist"),
		"--store", storeDir,
	})
	if err == nil {
		t.Fatal("runSmartWatcher() error = nil, want non-zero publication failure")
	}
	if _, err := os.Stat(filepath.Join(storeDir, "index.json")); err != nil {
		t.Fatalf("snapshot index not written despite export failure: %v", err)
	}
}
