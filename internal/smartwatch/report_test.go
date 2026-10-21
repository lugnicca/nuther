package smartwatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"nuther/internal/smart"
)

func reportSnapshotAt(device, serial string, now time.Time) Snapshot {
	drive := testDrive(device, serial)
	record := SnapshotRecord{
		ID:        "id-" + serial,
		Timestamp: now,
		Reason:    ReasonManual,
		Device:    SummaryForDrive(drive),
	}
	return ReportSnapshot(record, drive)
}

func reportLines(t *testing.T, path string) []Snapshot {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil
	}
	var snapshots []Snapshot
	for _, line := range strings.Split(trimmed, "\n") {
		var snapshot Snapshot
		if err := json.Unmarshal([]byte(line), &snapshot); err != nil {
			t.Fatalf("report line is not valid JSON: %v\n%s", err, line)
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

func TestAppendReport(t *testing.T) {
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	one := []Snapshot{reportSnapshotAt("/dev/sda", "SN-A", now)}
	two := []Snapshot{
		reportSnapshotAt("/dev/sda", "SN-A", now),
		reportSnapshotAt("/dev/sdb", "SN-B", now),
	}
	existing := []byte("{\"id\":\"old\"}\n")

	tests := []struct {
		name      string
		existing  []byte
		snapshots []Snapshot
		wantLines int
	}{
		{name: "one drive creates the file", snapshots: one, wantLines: 1},
		{name: "several drives append one line each", snapshots: two, wantLines: 2},
		{name: "nothing measured writes nothing", snapshots: nil, wantLines: 0},
		{name: "appends to an existing file", existing: existing, snapshots: one, wantLines: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "report.jsonl")
			if tt.existing != nil {
				if err := os.WriteFile(path, tt.existing, 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			}
			if err := AppendReport(path, tt.snapshots); err != nil {
				t.Fatalf("AppendReport() error = %v", err)
			}
			if tt.wantLines == 0 {
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("file created for an empty append: %v", err)
				}
				return
			}
			lines := reportLines(t, path)
			if len(lines) != tt.wantLines {
				t.Fatalf("report lines = %d, want %d", len(lines), tt.wantLines)
			}
			for i, want := range tt.snapshots {
				got := lines[len(lines)-len(tt.snapshots)+i]
				if got.ID != want.ID || got.SMART.Serial != want.SMART.Serial {
					t.Fatalf("line %d = %s/%s, want %s/%s", i, got.ID, got.SMART.Serial, want.ID, want.SMART.Serial)
				}
			}
		})
	}
}

func TestAppendReportRotatesAtCap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.jsonl")
	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), ReportLogMaxBytes), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	snapshot := reportSnapshotAt("/dev/sda", "SN-A", time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC))

	if err := AppendReport(path, []Snapshot{snapshot}); err != nil {
		t.Fatalf("AppendReport() error = %v", err)
	}
	rotated, err := os.Stat(path + ".1")
	if err != nil {
		t.Fatalf("rotated file missing: %v", err)
	}
	if rotated.Size() != ReportLogMaxBytes {
		t.Fatalf("rotated size = %d, want %d", rotated.Size(), ReportLogMaxBytes)
	}
	lines := reportLines(t, path)
	if len(lines) != 1 || lines[0].SMART.Serial != "SN-A" {
		t.Fatalf("active file after rotation = %+v, want one SN-A line", lines)
	}
}

// The report line must be the store's own Snapshot object, not a second shape
// that can drift from what the archive holds.
func TestAppendReportLineMatchesStoredSnapshot(t *testing.T) {
	store := NewStore(t.TempDir())
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	drive := testDrive("/dev/sda", "SN-A")
	record, err := store.SaveSnapshot(now, ReasonManual, drive)
	if err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "report.jsonl")
	if err := AppendReport(path, []Snapshot{ReportSnapshot(record, drive)}); err != nil {
		t.Fatalf("AppendReport() error = %v", err)
	}
	stored, err := store.ReadSnapshot(record.ID)
	if err != nil {
		t.Fatalf("ReadSnapshot() error = %v", err)
	}
	lines := reportLines(t, path)
	if len(lines) != 1 {
		t.Fatalf("report lines = %d, want 1", len(lines))
	}
	if !reflect.DeepEqual(lines[0], stored) {
		t.Fatalf("report line differs from stored snapshot\n got: %+v\nwant: %+v", lines[0], stored)
	}
}

// A scan that measured nothing must append nothing, and the failed scan must
// still return an error.
func TestWatcherSnapshotOnceNoReportLineWhenNothingMeasured(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.jsonl")
	watcher := NewWatcher(WatcherConfig{
		Store: NewStore(t.TempDir()),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{failedDrive("/dev/sda")}, nil
		},
		ReportLog: path,
	})
	if _, err := watcher.SnapshotOnce(t.Context(), ReasonManual); err == nil {
		t.Fatal("SnapshotOnce() error = nil, want a failed scan")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("report log written for a scan that measured nothing: %v", err)
	}
}

// The flag path must append exactly one line per measured drive, including the
// full S.M.A.R.T. report, and skip a drive whose own scan failed.
func TestWatcherSnapshotOnceAppendsOneReportLinePerMeasuredDrive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.jsonl")
	watcher := NewWatcher(WatcherConfig{
		Store: NewStore(t.TempDir()),
		Hub:   NewEventHub(""),
		Scanner: func() ([]smart.DriveInfo, error) {
			return []smart.DriveInfo{
				testDrive("/dev/sda", "SN-A"),
				testDrive("/dev/sdb", "SN-B"),
				failedDrive("/dev/sdc"),
			}, nil
		},
		Now:       func() time.Time { return time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC) },
		ReportLog: path,
	})
	records, err := watcher.SnapshotOnce(t.Context(), ReasonManual)
	if err != nil {
		t.Fatalf("SnapshotOnce() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
	lines := reportLines(t, path)
	if len(lines) != 2 {
		t.Fatalf("report lines = %d, want one per measured drive", len(lines))
	}
	got := map[string]bool{}
	for _, line := range lines {
		if line.Reason != ReasonManual || line.Timestamp.IsZero() || line.Device.Key == "" {
			t.Fatalf("report line missing snapshot fields: %+v", line)
		}
		got[line.SMART.Serial] = true
	}
	if !got["SN-A"] || !got["SN-B"] {
		t.Fatalf("report serials = %v, want SN-A and SN-B", got)
	}
}
