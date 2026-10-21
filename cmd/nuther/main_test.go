package main

import (
	"os"
	"path/filepath"
	"strings"
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

// A --report-log append that fails must still archive the snapshot and must
// still exit non-zero.
func TestRunSmartWatcherOnceReportFailureStillArchives(t *testing.T) {
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
		"--store", storeDir,
		"--report-log", filepath.Join(t.TempDir(), "missing", "report.jsonl"),
	})
	if err == nil {
		t.Fatal("runSmartWatcher() error = nil, want non-zero report failure")
	}
	if _, err := os.Stat(filepath.Join(storeDir, "index.json")); err != nil {
		t.Fatalf("snapshot index not written despite report failure: %v", err)
	}
}

// --report-log appends exactly one line per measured drive, so the fleet map
// can read the full report from the newest line.
func TestRunSmartWatcherOnceWritesOneReportLinePerMeasuredDrive(t *testing.T) {
	previous := scanDrives
	scanDrives = func() ([]smart.DriveInfo, error) {
		return []smart.DriveInfo{
			{Device: "/dev/sda", Model: "Test Drive A", Serial: "SN-A", Interface: "SATA SSD", HealthStatus: smart.HealthGood, HealthPassed: true},
			{Device: "/dev/sdb", Model: "Test Drive B", Serial: "SN-B", Interface: "SATA SSD", HealthStatus: smart.HealthGood, HealthPassed: true},
		}, nil
	}
	defer func() { scanDrives = previous }()

	reportPath := filepath.Join(t.TempDir(), "report.jsonl")
	err := runSmartWatcher("watch-smart", []string{
		"--once",
		"--store", t.TempDir(),
		"--report-log", reportPath,
	})
	if err != nil {
		t.Fatalf("runSmartWatcher() error = %v", err)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("report lines = %d, want 2", len(lines))
	}
	for _, line := range lines {
		for _, want := range []string{`"id":`, `"timestamp":`, `"reason":`, `"device":`, `"smart":`} {
			if !strings.Contains(line, want) {
				t.Fatalf("report line missing %s: %s", want, line)
			}
		}
	}
}
